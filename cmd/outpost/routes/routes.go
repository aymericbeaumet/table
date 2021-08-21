package routes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jointable/table/pkg/ecdsa"
	"github.com/jointable/table/pkg/ip"
	"github.com/jointable/table/pkg/jwt"
	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	log "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

/*
 * `outpost` will one day become a standalone service. This will become
 * necessary when we scale as it will give us a better control over our
 * resources, and make the maintenance easier (the api is stateless, outpost is
 * stateful).
 *
 * For now for the sake of simplicity it is embedded inside the API service,
 * which will always be the main entrypoint. That being said, we isolate all
 * the outpost code in this directory to minimize the cost of splitting the
 * services in the future.
 */

func Up(r *gin.Engine) error {
	// Compute ballast

	ballast, err := getBallastMiB()
	if err != nil {
		panic(err)
	}

	// Infer nat1To1IP

	nat1To1IP, err := getNAT1TO1IP(context.Background())
	if err != nil {
		panic(err)
	}

	// Prepare SFU (https://github.com/pion/ion-sfu/blob/master/config.toml)

	logger := log.New()
	sfu.Logger = logger

	conf := sfu.Config{}
	conf.Router.AudioLevelFilter = 20
	conf.Router.AudioLevelInterval = 250
	conf.Router.AudioLevelThreshold = 35
	conf.Router.MaxBandwidth = 1500
	conf.Router.MaxPacketTrack = 500
	conf.Router.Simulcast.BestQualityFirst = true
	conf.Router.Simulcast.EnableTemporalLayer = true
	conf.Router.WithStats = true
	conf.SFU.Ballast = ballast
	conf.SFU.WithStats = true
	conf.WebRTC.Candidates.IceLite = true
	conf.WebRTC.Candidates.NAT1To1IPs = []string{nat1To1IP}
	conf.WebRTC.ICEPortRange = []uint16{40000, 49999}
	conf.WebRTC.SDPSemantics = "unified-plan"
	conf.WebRTC.Timeouts.ICEDisconnectedTimeout = 5
	conf.WebRTC.Timeouts.ICEFailedTimeout = 25
	conf.WebRTC.Timeouts.ICEKeepaliveInterval = 2

	fmt.Printf("SFU conf %#v\n", conf)

	s := sfu.NewSFU(conf)

	apiDC := s.NewDatachannel(sfu.APIChannelLabel)
	apiDC.Use(datachannel.SubscriberAPI)

	attachBroadcastChannel(s, "reliable")   // TODO: ordered=false
	attachBroadcastChannel(s, "unreliable") // TODO: ordered=false, maxRetransmits=0

	apiSigningPublicKey, err := ecdsa.LoadPublicKey(os.Getenv("API_SIGNING_PUBLIC_KEY"))
	if err != nil {
		return err
	}

	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/:token", func(c *gin.Context) {
		var params struct {
			Token string `uri:"token" bindings:"required,jwt"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		if _, err := jwt.ParseTableToken(params.Token, apiSigningPublicKey); err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer ws.Close()
		ws.EnableWriteCompression(true)

		peer := server.NewJSONSignal(sfu.NewPeer(s), logger)
		defer peer.Close()

		jc := jsonrpc2.NewConn(c, websocketjsonrpc2.NewObjectStream(ws), peer)
		<-jc.DisconnectNotify()
	})

	r.GET("/debug", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		sessionsByID := map[string]int{}
		for _, session := range s.GetSessions() {
			sessionsByID[session.ID()] = len(session.Peers())
		}

		c.JSON(http.StatusOK, gin.H{
			"runtime": struct {
				AllocMegabyte uint64 `json:"alloc_megabyte"`
				NumGoroutine  int    `json:"num_goroutine"`
			}{
				AllocMegabyte: m.Alloc / 1_000_000,
				NumGoroutine:  runtime.NumGoroutine(),
			},
			"sessions": sessionsByID,
		})
	})

	return nil
}

func getBallastMiB() (int64, error) {
	pctString := os.Getenv("WEBRTC_BALLAST_PCT")

	pct, err := strconv.ParseInt(pctString, 10, 64)
	if err != nil {
		return 0, err
	}

	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return int64(memory.Total) * pct / 100 / 1024 / 1024, nil
}

func getNAT1TO1IP(ctx context.Context) (string, error) {
	switch v := os.Getenv("WEBRTC_NAT1TO1IP"); v {
	case "local":
		return ip.GetLocalIP()
	case "public":
		return ip.GetPublicIP(ctx)
	default:
		return "", fmt.Errorf("implementation error: unexpected WEBRTC_NAT1TO1IP %#v", v)
	}
}

func attachBroadcastChannel(s *sfu.SFU, label string) {
	dc := s.NewDatachannel(label)
	dc.OnMessage(func(ctx context.Context, args sfu.ProcessArgs) {
		p := args.Peer.Publisher()
		if p == nil {
			return
		}

		tracks := p.Tracks()
		if len(tracks) == 0 {
			fmt.Println("warning: missing tracks for peer:", p)
			return
		}

		data, err := msgpack.Marshal(struct {
			StreamID string             `msgpack:"streamID"`
			Payload  msgpack.RawMessage `msgpack:"payload"`
		}{
			StreamID: tracks[0].StreamID(),
			Payload:  args.Message.Data,
		})
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		for _, dc := range args.Peer.Session().GetDataChannels("", label) {
			if dc != args.DataChannel {
				_ = dc.Send(data)
			}
		}
	})
}

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	log "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	// Load env

	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env." + mode)

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

	// HTTP server

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	r.Use(cors.New(config))

	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/tables/:name", func(c *gin.Context) {
		var params struct {
			SessionName string `uri:"name"`
		}
		if c.BindUri(&params) != nil {
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

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", os.Getenv("HOSTNAME"), os.Getenv("PORT")),
		Handler: r,
	}

	// If possible, run with HTTPS + check origin certificate
	if cf, kf := os.Getenv("CERT_FILE"), os.Getenv("KEY_FILE"); len(cf) > 0 && len(kf) > 0 {
		origin, err := os.ReadFile(os.Getenv("ORIGIN_CERT_FILE"))
		if err != nil {
			panic(err)
		}

		clientsCAs := x509.NewCertPool()
		if !clientsCAs.AppendCertsFromPEM(origin) {
			panic(errors.New("failed to parse origin certificate"))
		}

		server.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  clientsCAs,
		}

		fmt.Printf("Listening on https://%s\n", server.Addr)
		panic(server.ListenAndServeTLS(cf, kf))
	}

	// Otherwise run with HTTP
	fmt.Printf("Listening on http://%s\n", server.Addr)
	panic(server.ListenAndServe())
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
		return GetLocalIP()
	case "public":
		return GetPublicIP(ctx)
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

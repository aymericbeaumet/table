package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	log "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

type DataChannelEnvelope struct {
	Seq      uint32             `msgpack:"seq"`
	StreamID string             `msgpack:"streamID"`
	Payload  msgpack.RawMessage `msgpack:"payload"`
}

func attachBroadcastChannel(s *sfu.SFU, label string) {
	var seq uint32

	dc := s.NewDatachannel(label)
	dc.OnMessage(func(ctx context.Context, args sfu.ProcessArgs) {
		p := args.Peer.Publisher()
		if p == nil {
			return
		}

		tracks := p.Tracks()
		if len(tracks) == 0 {
			return
		}

		data, err := msgpack.Marshal(DataChannelEnvelope{
			Seq:      atomic.AddUint32(&seq, 1),
			StreamID: tracks[0].StreamID(),
			Payload:  args.Message.Data,
		})
		if err != nil {
			return
		}

		for _, dc := range args.Peer.Session().GetDataChannels("", label) {
			if dc != args.DataChannel {
				_ = dc.Send(data)
			}
		}
	})
}

func main() {
	// Load env

	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env." + mode)

	// Prepare SFU

	logger := log.New()
	sfu.Logger = logger

	conf := sfu.Config{}
	conf.SFU.Ballast = 0
	conf.SFU.WithStats = false
	conf.Router.MaxBandwidth = 1500
	conf.Router.MaxPacketTrack = 500
	conf.Router.AudioLevelThreshold = 40
	conf.Router.AudioLevelInterval = 1000
	conf.Router.AudioLevelFilter = 20
	conf.Router.Simulcast.BestQualityFirst = true
	conf.Router.Simulcast.EnableTemporalLayer = true
	conf.WebRTC.ICEPortRange = []uint16{5000, 5200}
	conf.WebRTC.ICEServers = []sfu.ICEServerConfig{
		{URLs: strings.Split(os.Getenv("STUN_SERVERS_URLS"), ",")},
	}
	conf.WebRTC.SDPSemantics = "unified-plan"
	conf.WebRTC.MDNS = true
	conf.WebRTC.Timeouts.ICEDisconnectedTimeout = 5
	conf.WebRTC.Timeouts.ICEFailedTimeout = 25
	conf.WebRTC.Timeouts.ICEKeepaliveInterval = 2
	conf.Turn.Enabled = false

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

	r.GET("/sessions/:name", func(c *gin.Context) {
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

	hostname := os.Getenv("HOSTNAME")
	if len(hostname) == 0 {
		hostname = "localhost"
	}
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	if cf, kf := os.Getenv("CERT_FILE"), os.Getenv("KEY_FILE"); len(cf) > 0 && len(kf) > 0 {
		panic(r.RunTLS(fmt.Sprintf("%s:%s", hostname, port), cf, kf))
	} else {
		panic(r.Run(fmt.Sprintf("%s:%s", hostname, port)))
	}
}

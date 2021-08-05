package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	log "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
)

type InboundSignalingMessage struct {
	SignalingMessageAnswer
	SignalingMessageCandidate
	SignalingMessageNegotiate
}

type SignalingMessageAnswer struct {
	Answer *webrtc.SessionDescription `json:"answer"`
}

type SignalingMessageCandidate struct {
	Candidate *webrtc.ICECandidateInit `json:"candidate"`
}

type SignalingMessageNegotiate struct {
	Negotiate *bool `json:"negotiate"`
}

type SignalingMessageOffer struct {
	Offer *webrtc.SessionDescription `json:"offer"`
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
	dc := s.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

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

	if port == "letsencrypt" {
		panic(autotls.Run(r, hostname))
	} else {
		panic(r.Run(fmt.Sprintf("%s:%s", hostname, port)))
	}
}

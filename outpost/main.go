package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/lucsky/cuid"
	"github.com/pion/ion-sfu/pkg/sfu"
	webrtc "github.com/pion/webrtc/v3"
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
	conf.WebRTC.SDPSemantics = "unified-plan"
	conf.WebRTC.MDNS = true
	conf.WebRTC.Timeouts.ICEDisconnectedTimeout = 5
	conf.WebRTC.Timeouts.ICEFailedTimeout = 25
	conf.WebRTC.Timeouts.ICEKeepaliveInterval = 2
	conf.Turn.Enabled = false
	s := sfu.NewSFU(conf)

	reliable := s.NewDatachannel("reliable")
	reliable.OnMessage(func(ctx context.Context, args sfu.ProcessArgs) {})

	unreliable := s.NewDatachannel("unreliable")
	unreliable.OnMessage(func(ctx context.Context, args sfu.ProcessArgs) {})

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

		var mu sync.Mutex
		writeJSON := func(v interface{}) error {
			mu.Lock()
			defer mu.Unlock()
			return ws.WriteJSON(v)
		}

		peerID := cuid.New()

		peer := sfu.NewPeer(s)
		defer peer.Close()

		peer.OnOffer = func(offer *webrtc.SessionDescription) {
			if err := writeJSON(&SignalingMessageOffer{Offer: offer}); err != nil {
				fmt.Println("error:", err)
				return
			}
		}

		peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, _ int) {
			if err := writeJSON(&SignalingMessageCandidate{Candidate: candidate}); err != nil {
				fmt.Println("error:", err)
				return
			}
		}

		if err := peer.Join(params.SessionName, peerID, sfu.JoinConfig{
			NoAutoSubscribe: false,
			NoPublish:       false,
			NoSubscribe:     false,
		}); err != nil {
			fmt.Println("error:", err)
			return
		}

		for {
			var message InboundSignalingMessage
			if err := ws.ReadJSON(&message); err != nil {
				if _, ok := err.(*websocket.CloseError); !ok {
					panic(err)
				}
				return
			}

			if message.Negotiate != nil {
				peer.Subscriber().Negotiate()
			} else if message.Answer != nil {
				if err = peer.SetRemoteDescription(*message.Answer); err != nil {
					fmt.Println("error:", err)
					return
				}
			} else {
				var candidate webrtc.ICECandidateInit
				if message.Candidate != nil {
					candidate = *message.Candidate
				}
				if err := peer.Trickle(candidate, 0); err != nil {
					panic(err)
				}
			}
		}
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
		log.Fatal(autotls.Run(r, hostname))
	} else {
		log.Fatal(r.Run(fmt.Sprintf("%s:%s", hostname, port)))
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aymericbeaumet/chill-room/outpost/sfu"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	webrtc "github.com/pion/webrtc/v3"
)

type InboundSignalingMessage struct {
	SignalingMessageCandidate
	SignalingMessageDescription
}

type SignalingMessageCandidate struct {
	Candidate *webrtc.ICECandidateInit `json:"candidate"`
}

type SignalingMessageDescription struct {
	Description *webrtc.SessionDescription `json:"description"`
}

type SignalingMessageNegotiationNeeded struct {
	NegotiationNeeded bool `json:"negotiationneeded"`
}

func main() {
	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	godotenv.Load(".env")
	godotenv.Load(".env." + mode)

	s := sfu.New()

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	r.Use(cors.New(config))

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	r.GET("/rooms/:name", func(c *gin.Context) {
		var params struct {
			Name string `uri:"name"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		room, err := s.UpsertRoom(params.Name)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		peer, err := sfu.NewPeer()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := room.RegisterPeer(peer); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer ws.Close()
		ws.EnableWriteCompression(true)

		var writeJSONmu sync.Mutex
		writeJSON := func(v interface{}) error {
			writeJSONmu.Lock()
			defer writeJSONmu.Unlock()
			return ws.WriteJSON(v)
		}

		peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			var ic *webrtc.ICECandidateInit
			if candidate != nil {
				j := candidate.ToJSON()
				ic = &j
			}

			if err := writeJSON(
				&SignalingMessageCandidate{Candidate: ic},
			); err != nil {
				panic(err)
			}
		})

		peer.OnNegotiationNeeded(func() {
			if err := writeJSON(
				&SignalingMessageNegotiationNeeded{NegotiationNeeded: true},
			); err != nil {
				panic(err)
			}
		})

		for {
			var message InboundSignalingMessage
			if err := ws.ReadJSON(&message); err != nil {
				if _, ok := err.(*websocket.CloseError); !ok {
					panic(err)
				}
				return
			}

			if message.Description != nil {
				localDesc, err := peer.Handshake(*message.Description)
				if err != nil {
					panic(err)
				}
				if err := writeJSON(
					&SignalingMessageDescription{Description: localDesc},
				); err != nil {
					panic(err)
				}
			} else {
				var c webrtc.ICECandidateInit
				if message.Candidate != nil {
					c = *message.Candidate
				}
				if err := peer.AddICECandidate(c); err != nil {
					panic(err)
				}
			}
		}
	})

	r.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"sfu": s,
			"runtime": map[string]interface{}{
				"goroutines": runtime.NumGoroutine(),
			},
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

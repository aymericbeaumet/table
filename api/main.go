package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	webrtc "github.com/pion/webrtc/v3"
)

// The same signaling message is used both ways
type SignalingMessage struct {
	ANSWER        *webrtc.SessionDescription `json:"ANSWER,omitempty"`
	ICE_CANDIDATE *webrtc.ICECandidateInit   `json:"ICE_CANDIDATE,omitempty"`
	OFFER         *webrtc.SessionDescription `json:"OFFER,omitempty"`
}

func main() {
	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	godotenv.Load(".env")
	godotenv.Load(".env." + mode)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	rooms := NewRooms()

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	r.Use(cors.New(config))

	r.GET("/rooms/:name", func(c *gin.Context) {
		var params struct {
			Name string `uri:"name"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		room := rooms.Upsert(params.Name)

		peer, err := NewPeer(room.Inc())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer peer.Close()

		peer.OnOpen(func() {
			if room.AddPeer(peer) {
				room.EachPeer(func(p *Peer) {
					if p != peer {
						p.SendPeer(peer)
						peer.SendPeer(p)
					}
				})
			}
		})

		peer.OnTrack(func(track webrtc.TrackLocal) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					p.AddLocalTrack(track)
				}
			})
		})

		peer.OnMessageReliable(func(msg []byte) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					if err := p.SendReliable(peer, msg); err != nil {
						fmt.Println("error:", err)
					}
				}
			})
		})

		peer.OnMessageUnreliable(func(msg []byte) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					if err := p.SendUnreliable(peer, msg); err != nil {
						fmt.Println("error:", err)
					}
				}
			})
		})

		peer.OnClose(func() {
			if room.DelPeer(peer) {
				room.EachPeer(func(p *Peer) {
					p.SendPeer(peer)
				})
			}
		})

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer ws.Close()
		ws.EnableWriteCompression(true)

		peer.OnOffer(func(offer webrtc.SessionDescription) {
			ws.WriteJSON(&SignalingMessage{OFFER: &offer})
		})

		peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			j := candidate.ToJSON()
			ws.WriteJSON(&SignalingMessage{ICE_CANDIDATE: &j})
		})

		for {
			var message SignalingMessage
			if err := ws.ReadJSON(&message); err != nil {
				fmt.Println("error:", err)
				continue
			}

			if message.ICE_CANDIDATE != nil {
				peer.AddRemoteICECandidate(*message.ICE_CANDIDATE)
			}

			if message.OFFER != nil {
				answer, err := peer.AddRemoteOffer(*message.OFFER)
				if err != nil {
					fmt.Println("error:", err)
					continue
				}
				ws.WriteJSON(&SignalingMessage{ANSWER: answer})
			}
		}
	})

	r.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"roomsByID": rooms,
		})
	})

	hostname := os.Getenv("HOSTNAME")
	if len(hostname) == 0 {
		hostname = "0.0.0.0"
	}
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(r.Run(fmt.Sprintf("%s:%s", hostname, port)))
}

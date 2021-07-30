package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	webrtc "github.com/pion/webrtc/v3"
)

type InboundSignalingMessage struct {
	SignalingMessageDescription
	SignalingMessageIceCandidate
}

type SignalingMessageDescription struct {
	Description *webrtc.SessionDescription `json:"description"`
}

type SignalingMessageIceCandidate struct {
	Candidate *webrtc.ICECandidateInit `json:"candidate"`
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

	rooms := NewRooms()

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

		room := rooms.Upsert(params.Name)

		peer, err := NewPeer(room.Inc())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer peer.Close()

		peer.OnConnected(func() {
			if room.AddPeer(peer) {
				room.EachPeer(func(p *Peer) {
					p.EachTrack(func(t webrtc.TrackLocal) {
						if err := peer.AddTrack(t); err != nil {
							fmt.Println("error:", err)
						}
					})
				})
			}
		})

		peer.OnTrack(func(track webrtc.TrackLocal) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					if err := p.AddTrack(track); err != nil {
						fmt.Println("error:", err)
					}
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

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer ws.Close()
		ws.EnableWriteCompression(true)

		// TODO: find a cleaner way to prevent concurrent writes
		var mu sync.Mutex
		writeJSON := func(v interface{}) error {
			mu.Lock()
			defer mu.Unlock()
			return ws.WriteJSON(v)
		}

		peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			var ic *webrtc.ICECandidateInit
			if candidate != nil {
				j := candidate.ToJSON()
				ic = &j
			}
			writeJSON(&SignalingMessageIceCandidate{Candidate: ic})
		})

		peer.OnNegotiationNeeded(func() {
			writeJSON(&SignalingMessageNegotiationNeeded{NegotiationNeeded: true})
		})

		for {
			var message InboundSignalingMessage
			if err := ws.ReadJSON(&message); err != nil {
				fmt.Println("error:", err)
				return
			}

			if message.Description != nil {
				desc := message.Description

				if desc.Type != webrtc.SDPTypeOffer {
					panic(fmt.Errorf("unexpected session description of type %#v", desc.Type))
				}

				if err := peer.peerConnection.SetRemoteDescription(*desc); err != nil {
					panic(err)
				}

				answer, err := peer.peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}

				if err := peer.peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}

				if err := writeJSON(&SignalingMessageDescription{Description: peer.LocalDescription()}); err != nil {
					panic(err)
				}
			}

			if message.Candidate != nil {
				if err := peer.peerConnection.AddICECandidate(*message.Candidate); err != nil {
					panic(err)
				}
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

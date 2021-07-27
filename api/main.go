package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	webrtc "github.com/pion/webrtc/v3"
)

type SignalingMessage struct {
	SignalingMessageDescription
	SignalingMessageIceCandidate
}

type SignalingMessageDescription struct {
	Description *webrtc.SessionDescription `json:"description"`
}

type SignalingMessageIceCandidate struct {
	Candidate *webrtc.ICECandidateInit `json:"candidate"`
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
					p.AddTrack(track)
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

		// TODO: find a cleaner way to prevent concurrent writes
		var mu sync.Mutex
		writeJSON := func(v interface{}) error {
			mu.Lock()
			defer mu.Unlock()
			return ws.WriteJSON(v)
		}

		/*
		 * https://w3c.github.io/webrtc-pc/#perfect-negotiation-example
		 * The perfect negotiation logic, separated from the rest of the application
		 */

		const polite = false

		// keep track of some negotiation state to prevent races and errors
		makingOffer := false
		ignoreOffer := false
		isSettingRemoteAnswerPending := false

		peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			var ic *webrtc.ICECandidateInit
			if candidate != nil {
				j := candidate.ToJSON()
				ic = &j
			}
			writeJSON(&SignalingMessageIceCandidate{Candidate: ic})
		})

		peer.OnNegotiationNeeded(func() {
			defer func() { makingOffer = false }()
			makingOffer = true
			offer, err := peer.peerConnection.CreateOffer(nil)
			if err != nil {
				panic(err)
			}
			peer.peerConnection.SetLocalDescription(offer)
			writeJSON(&SignalingMessageDescription{Description: peer.peerConnection.LocalDescription()})
		})

		for {
			var message SignalingMessage
			if err := ws.ReadJSON(&message); err != nil {
				fmt.Println("error:", err)
				return
			}

			if message.Description != nil {
				// An offer may come in while we are busy processing SRD(answer).
				// In this case, we will be in "stable" by the time the offer is processed
				// so it is safe to chain it on our Operations Chain now.
				readyForOffer :=
					!makingOffer &&
						(peer.peerConnection.SignalingState() == webrtc.SignalingStateStable || isSettingRemoteAnswerPending)
				offerCollision := message.Description.Type == webrtc.SDPTypeOffer && !readyForOffer

				ignoreOffer := !polite && offerCollision
				if ignoreOffer {
					return
				}

				isSettingRemoteAnswerPending = message.Description.Type == webrtc.SDPTypeAnswer
				peer.peerConnection.SetRemoteDescription(*message.Description)
				isSettingRemoteAnswerPending = false
				if message.Description.Type == webrtc.SDPTypeOffer {
					answer, _ := peer.peerConnection.CreateAnswer(nil)
					peer.peerConnection.SetLocalDescription(answer)
					writeJSON(&SignalingMessageDescription{Description: peer.LocalDescription()})
				}
			} else if message.Candidate != nil {
				if err := peer.peerConnection.AddICECandidate(*message.Candidate); err != nil && !ignoreOffer {
					fmt.Println("error:", err)
					return
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
		hostname = "0.0.0.0"
	}
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	log.Fatal(r.Run(fmt.Sprintf("%s:%s", hostname, port)))
}

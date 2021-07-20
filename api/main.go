package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	webrtc "github.com/pion/webrtc/v3"
)

type InboundMessage struct {
	Offer *InboundMessageOffer `json:"OFFER,omitempty"`
}

type InboundMessageOffer struct {
	webrtc.SessionDescription
}

type OutboundMessage struct {
	Answer *OutboundMessageAnswer `json:"ANSWER,omitempty"`
}

type OutboundMessageAnswer struct {
	webrtc.SessionDescription
}

func main() {
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	upgrader := websocket.Upgrader{
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: true,
	}

	r := gin.Default()

	r.GET("/rooms/:name", func(c *gin.Context) {
		//name, err := url.QueryUnescape(c.Param("name"))
		//if err != nil {
		//c.AbortWithError(http.StatusBadRequest, err)
		//return
		//}

		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()
		conn.EnableWriteCompression(true)

		// Create a new RTCPeerConnection
		peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
		if err != nil {
			panic(err)
		}
		defer peerConnection.Close()

		// Listen for ICE candidates
		var candidate *webrtc.ICECandidate
		peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
			if candidate != nil {
				return
			}

			peerConnection.AddICECandidate(c.ToJSON())

			conn.WriteJSON(&OutboundMessage{
				Answer: &OutboundMessageAnswer{
					SessionDescription: *peerConnection.LocalDescription(),
				},
			})

			candidate = c
		})

		for {
			var message InboundMessage
			if err := conn.ReadJSON(&message); err != nil {
				panic(err)
			}

			if m := message.Offer; m != nil {
				// Set the remote SessionDescription
				if err = peerConnection.SetRemoteDescription(message.Offer.SessionDescription); err != nil {
					panic(err)
				}

				// Create answer
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}

				// Sets the LocalDescription, and starts our UDP listeners
				if err = peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}
			}
		}
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

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebRTCSignal struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}

type InboundMessage struct {
	Signal *InboundMessageSignal `json:"SIGNAL,omitempty"`
}

type InboundMessageSignal struct {
	PeerID uuid.UUID    `json:"peerID"`
	Signal WebRTCSignal `json:"signal"`
}

type OutboundMessage struct {
	ExistingPeer   *OutboundMessageExistingPeer   `json:"EXISTING_PEER,omitempty"`
	SignalReceived *OutboundMessageSignalReceived `json:"SIGNAL_RECEIVED,omitempty"`
}

type OutboundMessageExistingPeer struct {
	PeerID uuid.UUID `json:"peerID"`
}

type OutboundMessageSignalReceived struct {
	PeerID uuid.UUID    `json:"peerID"`
	Signal WebRTCSignal `json:"signal"`
}

func main() {
	rooms := NewRooms()

	upgrader := websocket.Upgrader{
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: true,
	}

	r := gin.Default()

	r.GET("/rooms/:name", func(c *gin.Context) {
		name, err := url.QueryUnescape(c.Param("name"))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Make sure the room exists
		room := rooms.Upsert(name, 3)

		// Create the peer
		peer := NewPeer(c.Request.Header.Get("user-agent"), c.ClientIP())
		defer peer.Close()

		// Append the peer
		if err := room.AddPeer(peer); err != nil {
			c.AbortWithError(http.StatusServiceUnavailable, err)
			return
		}
		defer room.DelPeer(peer)

		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()
		conn.EnableWriteCompression(true)

		go peer.WriteLoop(conn)

		// Send all the existing peers to the new peer
		room.EachPeer(func(p *Peer) {
			if p.ID == peer.ID {
				return
			}
			peer.Write(&OutboundMessage{
				ExistingPeer: &OutboundMessageExistingPeer{
					PeerID: p.ID,
				},
			})
		})

		for {
			message, err := peer.Read(conn)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			// Forward any signal to the corresponding peer
			if m := message.Signal; m != nil {
				if p := room.GetPeer(m.PeerID); p != nil {
					p.Write(&OutboundMessage{
						SignalReceived: &OutboundMessageSignalReceived{
							PeerID: peer.ID,
							Signal: m.Signal,
						},
					})
				}
			}
		}
	})

	r.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"roomsByName": rooms,
		})
	})

	log.Fatal(r.Run())
}

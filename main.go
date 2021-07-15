package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
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
	PeerStatus *OutboundMessagePeerStatus `json:"PEER_STATUS,omitempty"`
	PeerSignal *OutboundMessagePeerSignal `json:"PEER_SIGNAL,omitempty"`
}

type OutboundMessagePeerStatus struct {
	PeerID uuid.UUID `json:"peerID"`
	Status string    `json:"status"`
}

type OutboundMessagePeerSignal struct {
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

		// Notify the peer with the other online peers
		room.EachPeer(func(p *Peer) {
			if p.ID == peer.ID {
				return
			}
			peer.Write(&OutboundMessage{
				PeerStatus: &OutboundMessagePeerStatus{PeerID: p.ID, Status: "online"},
			})
		})

		// Notify the other peers when going offline
		defer func() {
			room.EachPeer(func(p *Peer) {
				if p.ID == peer.ID {
					return
				}
				p.Write(&OutboundMessage{
					PeerStatus: &OutboundMessagePeerStatus{PeerID: peer.ID, Status: "offline"},
				})
			})
		}()

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
						PeerSignal: &OutboundMessagePeerSignal{
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

	// dev only
	u, _ := url.Parse("http://localhost:1234")
	rp := httputil.NewSingleHostReverseProxy(u)
	r.NoRoute(func(c *gin.Context) {
		rp.ServeHTTP(c.Writer, c.Request)
	})

	log.Fatal(r.Run("0.0.0.0:8080"))
}

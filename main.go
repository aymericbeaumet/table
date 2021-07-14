package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type InboundMessage struct {
	Handshake *struct {
		PeerID uuid.UUID `json:"peerID"`
		Signal *struct {
			Type string `json:"type"`
			Sdp  string `json:"sdp"`
		} `json:"signal,omitempty"`
	} `json:"HANDSHAKE,omitempty"`
}

type OutboundMessage struct {
	Handshake map[string]interface{} `json:"HANDSHAKE,omitempty"`
	Peer      map[string]interface{} `json:"PEER,omitempty"`
}

func main() {
	rooms := NewRooms()

	upgrader := websocket.Upgrader{
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: true,
	}

	r := gin.Default()

	r.GET("/rooms/:name", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()
		conn.EnableWriteCompression(true)

		peer := NewPeer(conn, c.Request.Header.Get("user-agent"), c.ClientIP())
		defer peer.Close()
		go peer.Loop()

		room := rooms.Upsert(c.Param("name"))

		// Send all the existing peers to the new peer
		room.EachPeer(func(p *Peer) {
			peer.Write(&OutboundMessage{
				Peer: map[string]interface{}{
					"peerID": p.ID,
				},
			})
		})

		// Append the new peer
		room.AddPeer(peer)
		defer room.DelPeer(peer)

		for {
			var message InboundMessage
			if err := peer.Read(&message); err != nil {
				fmt.Println("error reading message:", err)
				continue
			}

			// Forward the handshake request to the corresponding peer
			if m := message.Handshake; m != nil {
				room.GetPeer(m.PeerID).Write(&OutboundMessage{
					Handshake: map[string]interface{}{
						"peerID": peer.ID,
						"signal": m.Signal,
					},
				})
			}
		}
	})

	r.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"rooms": rooms,
		})
	})

	log.Fatal(r.Run())
}

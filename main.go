package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Handshake *struct {
		PeerID int `json:"peerID"`
		Signal struct {
			Type string `json:"type"`
			Sdp  string `json:"sdp"`
		} `json:"signal"`
	} `json:"HANDSHAKE"`
}

type Reply struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func main() {
	rooms := NewRooms()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	r := gin.Default()

	r.GET("/rooms/:name", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()

		room := rooms.Upsert(c.Param("name"))

		// Send all the existing peers to the new peer
		room.EachConn(func(i int, c *websocket.Conn) {
			conn.WriteJSON(&Reply{
				Type: "PEER",
				Payload: map[string]interface{}{
					"peerID": i,
				},
			})
		})

		// Append the new peer
		id := room.AddConn(conn)
		defer room.DelConn(id)

		for {
			var message Message
			if err := conn.ReadJSON(&message); err != nil {
				fmt.Println("error reading message:", err)
				continue
			}

			if m := message.Handshake; m != nil {
				c, ok := room.GetConn(m.PeerID)
				if !ok {
					fmt.Println("error missing peer")
					continue
				}

				c.WriteJSON(&Reply{
					Type: "HANDSHAKE",
					Payload: map[string]interface{}{
						"peerID": id,
						"signal": m.Signal,
					},
				})
			}
		}
	})

	r.GET("/debug", func(c *gin.Context) {
		c.JSON(http.StatusOK, rooms)
	})

	log.Fatal(r.Run())
}

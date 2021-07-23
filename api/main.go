package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	webrtc "github.com/pion/webrtc/v3"
)

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

	r.POST("/rooms", func(c *gin.Context) {
		var payload struct {
			Name  string                    `json:"name"`
			Offer webrtc.SessionDescription `json:"offer"`
		}
		if c.BindJSON(&payload) != nil {
			return
		}

		// Upsert room
		room := rooms.Upsert(payload.Name)

		// Create and init peer

		peer := NewPeer()

		if err := peer.Init(payload.Offer); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		peer.OnReliable(func(msg []byte) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					if err := p.SendReliable(msg); err != nil {
						fmt.Println("error:", err)
					}
				}
			})
		})

		peer.OnUnreliable(func(msg []byte) {
			room.EachPeer(func(p *Peer) {
				if p != peer {
					if err := p.SendUnreliable(msg); err != nil {
						fmt.Println("error:", err)
					}
				}
			})
		})

		peer.OnClose(func() {
			room.DelPeer(peer)
		})

		// Add peer
		room.AddPeer(peer)

		// Send the answer back
		c.JSON(http.StatusOK, struct {
			Answer *webrtc.SessionDescription `json:"answer"`
		}{
			Answer: peer.LocalDescription(),
		})
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

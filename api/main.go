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

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

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

		// Create a new RTCPeerConnection
		peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Set remote description
		if err = peerConnection.SetRemoteDescription(payload.Offer); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Sets the LocalDescription, and starts our UDP listeners
		if err = peerConnection.SetLocalDescription(answer); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Listen for data channels
		peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
			dc.OnOpen(func() {
				dc.SendText("Hello, World!")
			})
		})

		// Listen for ICE candidates
		done := make(chan struct{})
		peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				peerConnection.AddICECandidate(candidate.ToJSON())
			} else {
				done <- struct{}{}
			}
		})
		<-done

		// Send the answer back
		c.JSON(http.StatusOK, struct {
			Answer *webrtc.SessionDescription `json:"answer"`
		}{
			Answer: peerConnection.LocalDescription(),
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

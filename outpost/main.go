package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	log "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
)

var logger = log.New()

func main() {
	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	godotenv.Load(".env")
	godotenv.Load(".env." + mode)

	var c sfu.Config
	c.SFU.Ballast = 0
	c.SFU.WithStats = false
	c.Router.MaxBandwidth = 1500
	c.Router.MaxPacketTrack = 500
	c.Router.AudioLevelThreshold = 40
	c.Router.AudioLevelInterval = 1000
	c.Router.AudioLevelFilter = 20
	c.Router.Simulcast.BestQualityFirst = true
	c.Router.Simulcast.EnableTemporalLayer = false
	c.WebRTC.ICEPortRange = []uint16{5000, 5200}
	c.WebRTC.SDPSemantics = "unified-plan"
	c.WebRTC.MDNS = true
	c.WebRTC.Timeouts.ICEDisconnectedTimeout = 5
	c.WebRTC.Timeouts.ICEFailedTimeout = 25
	c.WebRTC.Timeouts.ICEKeepaliveInterval = 2
	c.Turn.Enabled = false

	s := sfu.NewSFU(c)

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	r.Use(cors.New(config))

	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/rooms/:name", func(c *gin.Context) {
		var params struct {
			Name string `uri:"name"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			panic(err)
		}
		defer ws.Close()

		p := server.NewJSONSignal(sfu.NewPeer(s), logger)
		defer p.Close()

		jc := jsonrpc2.NewConn(c, websocketjsonrpc2.NewObjectStream(ws), p)
		<-jc.DisconnectNotify()
	})

	hostname := os.Getenv("HOSTNAME")
	if len(hostname) == 0 {
		hostname = "localhost"
	}
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	var err error
	if port == "letsencrypt" {
		err = autotls.Run(r, hostname)
	} else {
		err = r.Run(fmt.Sprintf("%s:%s", hostname, port))
	}
	if err != nil {
		panic(err)
	}
}

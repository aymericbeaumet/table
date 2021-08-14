package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/jointable/table/cmd/outpost"
	_ "github.com/slack-go/slack"
)

func main() {
	// Load env

	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env." + mode)

	// HTTP server

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	r.Use(cors.New(config))

	// TODO: split outpost into its own service
	if err := outpost.Up(r); err != nil {
		panic(err)
	}

	if err := Up(r); err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", os.Getenv("HOSTNAME"), os.Getenv("PORT")),
		Handler: r,
	}

	// If possible, run with HTTPS + check origin certificate
	if cf, kf := os.Getenv("CERT_FILE"), os.Getenv("KEY_FILE"); len(cf) > 0 || len(kf) > 0 {
		origin, err := os.ReadFile(os.Getenv("ORIGIN_CERT_FILE"))
		if err != nil {
			panic(err)
		}

		clientsCAs := x509.NewCertPool()
		if !clientsCAs.AppendCertsFromPEM(origin) {
			panic(errors.New("failed to parse origin certificate"))
		}

		server.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  clientsCAs,
		}

		fmt.Printf("Listening on https://%s\n", server.Addr)
		panic(server.ListenAndServeTLS(cf, kf))
	}

	// Otherwise run with HTTP
	fmt.Printf("Listening on http://%s\n", server.Addr)
	panic(server.ListenAndServe())
}

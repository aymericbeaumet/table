package service

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func New() *gin.Engine {
	mode := os.Getenv("MODE")
	if len(mode) == 0 {
		mode = "development"
	}
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env." + mode)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowCredentials: false,
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowMethods:     []string{"POST", "GET", "HEAD", "PUT", "PATCH", "DELETE"},
		AllowOrigins:     strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ","),
		MaxAge:           12 * time.Hour,
	}))

	return r
}

package routes

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	TableName string `json:"tbl"`
	jwt.StandardClaims
}

func Up(r *gin.Engine) error {
	signingPrivateKeyRaw, err := os.ReadFile(os.Getenv("SIGNING_PRIVATE_KEY"))
	if err != nil {
		return err
	}
	signingPrivateKeyBlock, _ := pem.Decode(signingPrivateKeyRaw)
	signingPrivateKey, err := x509.ParseECPrivateKey(signingPrivateKeyBlock.Bytes)
	if err != nil {
		return err
	}

	signingPublicKey, err := os.ReadFile(os.Getenv("SIGNING_PUBLIC_KEY"))
	if err != nil {
		return err
	}

	r.GET("/es256", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", signingPublicKey)
	})

	r.POST("/tables", func(c *gin.Context) {
		var payload struct {
			Name string `json:"name" binding:"required,min=3"`
		}
		if c.BindJSON(&payload) != nil {
			return
		}

		// TODO: check authorization to join the table before generating the token
		// TODO: upsert the table in database

		outpostBaseURL, err := getOutpostURL(payload.Name)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		claims := Claims{
			payload.Name,
			jwt.StandardClaims{
				Audience:  outpostBaseURL,
				ExpiresAt: time.Now().Add(60 * time.Second).UTC().Unix(),
			},
		}

		token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString(signingPrivateKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"token": token})
	})

	if err := upAuth(r); err != nil {
		return err
	}

	if err := upWebhooks(r); err != nil {
		return err
	}

	return nil
}

// TODO: update this when splitting the outpost out of the api service so
// that this is generated dynamically with hostnames that are not this
// api's public base url
func getOutpostURL(tableName string) (string, error) {
	return "ws" + os.Getenv("PUBLIC_BASE_URL")[4:], nil
}

package routes

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jointable/table/pkg/jwt"
)

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

	r.POST("/tables", func(c *gin.Context) {
		var payload struct {
			TableName string `json:"name" binding:"required,min=3"`
		}
		if c.BindJSON(&payload) != nil {
			return
		}

		// TODO: check authorization to join the table before generating the token
		// TODO: upsert the table in database

		outpostBaseURL, err := getOutpostURL(payload.TableName)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		token, err := jwt.NewTableToken(payload.TableName, outpostBaseURL).Sign(signingPrivateKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Data(http.StatusCreated, "text/plain", []byte(token))
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
	return os.Getenv("OUTPOST_BASE_URL"), nil
}

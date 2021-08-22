package routes

import (
	"github.com/gin-gonic/gin"
)

func Up(r *gin.Engine) error {
	if err := upAuth(r); err != nil {
		return err
	}

	if err := upWebhooks(r); err != nil {
		return err
	}

	return nil
}

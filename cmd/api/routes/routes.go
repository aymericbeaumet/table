package routes

import (
	"github.com/gin-gonic/gin"
)

func Up(r *gin.Engine) error {
	if err := upAuth(r); err != nil {
		return err
	}

	if err := upCommands(r); err != nil {
		return err
	}

	return nil
}

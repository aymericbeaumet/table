package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/slack-go/slack"
)

func Up(r *gin.Engine) error {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.Getenv("PUBLIC_BASE_URL")+"/auth/google/callback"),
	)

	r.GET("/auth/:provider", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Param("provider"))
		c.Request.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Param("provider"))
		c.Request.URL.RawQuery = q.Encode()

		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, user)
	})

	r.GET("/auth/:provider/logout", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Param("provider"))
		c.Request.URL.RawQuery = q.Encode()

		gothic.Logout(c.Writer, c.Request)
	})

	r.POST("/slack/command", func(c *gin.Context) {
		sv, err := slack.NewSecretsVerifier(c.Request.Header, os.Getenv("SLACK_SIGNING_SECRET"))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Request.Body = io.NopCloser(io.TeeReader(c.Request.Body, &sv))

		cmd, err := slack.SlashCommandParse(c.Request)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if err := sv.Ensure(); err != nil {
			_ = c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		switch cmd.Command {
		case "/table":

			switch cmd.Text {
			case "help", "-h", "--help":
				c.JSON(http.StatusOK, &slack.Msg{
					ResponseType: "ephemeral",
					Text:         "How to use the `/table` command?",
					Attachments: []slack.Attachment{
						{
							Text: "This command doesn't expect any parameter. Invoking it will post the link to this channel's table.",
						},
					},
				})

			default:
				text := fmt.Sprintf(
					"Join <@%s> in the <https://jointable.io/slack-%s-%s|channel's table>!",
					cmd.UserID, cmd.TeamDomain, cmd.ChannelName,
				)
				c.JSON(http.StatusOK, &slack.Msg{
					ResponseType: "in_channel",
					Text:         text,
				})
			}

		default:
			_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("unsupported command: %#v", cmd.Command))
		}
	})

	return nil
}

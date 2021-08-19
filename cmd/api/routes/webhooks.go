package routes

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func upWebhooks(r *gin.Engine) error {
	r.POST("/discord/interactions", func(c *gin.Context) {
		discordPublicKey, err := hex.DecodeString(os.Getenv("DISCORD_PUBLIC_KEY"))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if !discordgo.VerifyInteraction(c.Request, ed25519.PublicKey(discordPublicKey)) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var interaction discordgo.Interaction
		if c.BindJSON(&interaction) != nil {
			return
		}

		if interaction.Type == discordgo.InteractionPing {
			c.JSON(http.StatusOK, gin.H{"type": interaction.Type})
			return
		}

		if interaction.Type == discordgo.InteractionApplicationCommand {
			text := fmt.Sprintf(
				"Join <@%s> in the [channel's table](https://jointable.io/discord:%s/%s)!",
				interaction.Member.User.ID, interaction.GuildID, interaction.ChannelID,
			)

			c.JSON(http.StatusOK, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: text,
				},
			})
			return
		}

		c.AbortWithStatus(http.StatusBadRequest)
	})

	r.POST("/slack/commands/table", func(c *gin.Context) {
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

		switch cmd.Text {
		case "help", "-h", "--help":
			c.JSON(http.StatusOK, &slack.Msg{
				ResponseType: "ephemeral",
				Text:         "How to use the `/table` command?",
				Attachments: []slack.Attachment{
					{
						Text: "This command doesn't expect a parameter. Invoking it will post the link to this channel's table.",
					},
				},
			})

		default:
			text := fmt.Sprintf(
				"Join <@%s> in the <https://jointable.io/slack:%s/%s|channel's table>!",
				cmd.UserID, cmd.TeamDomain, cmd.ChannelName,
			)
			c.JSON(http.StatusOK, &slack.Msg{
				ResponseType: "in_channel",
				Text:         text,
			})
		}
	})

	return nil
}

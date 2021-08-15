package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	discordProvider "github.com/markbates/goth/providers/discord"
	githubProvider "github.com/markbates/goth/providers/github"
	googleProvider "github.com/markbates/goth/providers/google"
	microsoftonlineProvider "github.com/markbates/goth/providers/microsoftonline"
	"github.com/slack-go/slack"
)

func Up(r *gin.Engine) error {
	goth.UseProviders(
		discordProvider.New(
			os.Getenv("DISCORD_CLIENT_ID"),
			os.Getenv("DISCORD_CLIENT_SECRET"),
			os.Getenv("PUBLIC_BASE_URL")+"/auth/discord/callback",
			"guilds",
		),
		githubProvider.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			os.Getenv("PUBLIC_BASE_URL")+"/auth/github/callback",
			"read:org",
		),
		googleProvider.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("PUBLIC_BASE_URL")+"/auth/google/callback",
		),
		microsoftonlineProvider.New(
			os.Getenv("MICROSOFTONLINE_CLIENT_ID"),
			os.Getenv("MICROSOFTONLINE_CLIENT_SECRET"),
			os.Getenv("PUBLIC_BASE_URL")+"/auth/microsoftonline/callback",
		),
	)

	r.GET("/auth/:provider", func(c *gin.Context) {
		var params struct {
			Provider string `uri:"provider" binding:"required"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		q := c.Request.URL.Query()
		q.Set("provider", params.Provider)
		c.Request.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		var params struct {
			Provider string `uri:"provider" binding:"required"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		q := c.Request.URL.Query()
		q.Set("provider", params.Provider)
		c.Request.URL.RawQuery = q.Encode()

		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		authorizedScopes, err := getAuthorizedScopes(c.Request.Context(), user)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Some sanitization checks
		for _, s := range authorizedScopes {
			if strings.Count(s, ":") != 1 {
				_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("unexpected number of `:` in the scope %#v", s))
				return
			}
			if strings.Contains(s, "/") {
				_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("unexpected presence of `/` in the scope %#v", s))
				return
			}
		}

		c.JSON(http.StatusOK, authorizedScopes)
	})

	r.GET("/auth/:provider/logout", func(c *gin.Context) {
		var params struct {
			Provider string `uri:"provider" binding:"required"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		q := c.Request.URL.Query()
		q.Set("provider", params.Provider)
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

type DiscordUserGuilds []struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Owner       bool     `json:"owner"`
	Permissions uint64   `json:"permissions"`
	Features    []string `json:"features"`
}

type GithubUserOrgs []struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	URL              string `json:"url"`
	ReposURL         string `json:"repos_url"`
	EventsURL        string `json:"events_url"`
	HooksURL         string `json:"hooks_url"`
	IssuesURL        string `json:"issues_url"`
	MembersURL       string `json:"members_url"`
	PublicMembersURL string `json:"public_members_url"`
	AvatarURL        string `json:"avatar_url"`
	Description      string `json:"description"`
}

func getAuthorizedScopes(ctx context.Context, user goth.User) ([]string, error) {
	switch user.Provider {
	case "discord":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/users/@me/guilds", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+user.AccessToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var guilds DiscordUserGuilds
		if err := json.NewDecoder(res.Body).Decode(&guilds); err != nil {
			return nil, err
		}

		out := make([]string, 0, len(guilds))
		for _, guild := range guilds {
			out = append(out, fmt.Sprintf("discord:%s", guild.Name))
		}
		return out, nil

	case "google":
		splits := strings.Split(user.Email, "@")
		if len(splits) != 2 || len(splits[1]) == 0 {
			return nil, fmt.Errorf("malformated email %#v", user.Email)
		}
		return []string{
			fmt.Sprintf("gmail:%s", splits[1]),
			fmt.Sprintf("google:%s", splits[1]),
		}, nil

	case "github":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/orgs", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+user.AccessToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var orgs GithubUserOrgs
		if err := json.NewDecoder(res.Body).Decode(&orgs); err != nil {
			return nil, err
		}

		out := make([]string, 0, len(orgs))
		for _, org := range orgs {
			out = append(out, fmt.Sprintf("github:%s", org.Login))
		}
		return out, nil

	case "microsoftonline":
		splits := strings.Split(user.Email, "@")
		if len(splits) != 2 || len(splits[1]) == 0 {
			return nil, fmt.Errorf("malformated email %#v", user.Email)
		}
		return []string{
			fmt.Sprintf("microsoft:%s", splits[1]),
			fmt.Sprintf("outlook:%s", splits[1]),
		}, nil

	default:
		return nil, fmt.Errorf("implementation error: unsupported provider %#v", user.Provider)

	}
}

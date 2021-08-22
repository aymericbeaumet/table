package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jointable/table/pkg/ecdsa"
	"github.com/jointable/table/pkg/jwt"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	discordProvider "github.com/markbates/goth/providers/discord"
	githubProvider "github.com/markbates/goth/providers/github"
	googleProvider "github.com/markbates/goth/providers/google"
	microsoftonlineProvider "github.com/markbates/goth/providers/microsoftonline"
	slackProvider "github.com/markbates/goth/providers/slack"
)

func upAuth(r *gin.Engine) error {
	signingPrivateKey, err := ecdsa.LoadPrivateKey(os.Getenv("SIGNING_PRIVATE_KEY"))
	if err != nil {
		return err
	}

	store := sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SESSION_SECRET")))
	store.MaxAge(24 * 60 * 60)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = strings.HasPrefix(os.Getenv("PUBLIC_BASE_URL"), "https://")
	gothic.Store = store

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
		slackProvider.New(
			os.Getenv("SLACK_CLIENT_ID"),
			os.Getenv("SLACK_CLIENT_SECRET"),
			os.Getenv("PUBLIC_BASE_URL")+"/auth/slack/callback",
			"openid", "email", "profile",
		),
	)

	r.GET("/auth/table/:name", func(c *gin.Context) {
		var params struct {
			TableName string `uri:"name" binding:"required"`
		}
		if c.BindUri(&params) != nil {
			return
		}

		var query struct {
			Origin string `form:"origin" binding:"required"`
		}
		if c.BindQuery(&query) != nil {
			return
		}
		if query.Origin != os.Getenv("CORS_ALLOW_ORIGINS") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		scope, id, err := normalizeTableName(params.TableName)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var tableName string
		if len(scope) > 0 {
			tableName = scope + "/"
		}
		tableName += id

		// TODO: if authentication is needed, redirect to /auth/:provider
		token, err := jwt.NewTableToken(tableName).Sign(signingPrivateKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		next := fmt.Sprintf("%s/%s?token=%s", query.Origin, url.QueryEscape(tableName), url.QueryEscape(token))
		c.Redirect(http.StatusFound, next)
	})

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
		for _, scope := range authorizedScopes {
			if len(scope) == 0 {
				_ = c.AbortWithError(http.StatusInternalServerError, errors.New("unexpected empty scope"))
				return
			}

			if strings.Count(scope, ":") != 1 {
				_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("unexpected number of `:` in the scope %#v", scope))
				return
			}

			if strings.Contains(scope, "/") {
				_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("unexpected presence of `/` in the scope %#v", scope))
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

		_ = gothic.Logout(c.Writer, c.Request)
	})

	return nil
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

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code from the Discord API, got %#v", res.StatusCode)
		}

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

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code from the GitHub API, got %#v", res.StatusCode)
		}

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

	case "slack":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://slack.com/api/openid.connect.userInfo", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+user.AccessToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code from the Slack API, got %#v", res.StatusCode)
		}

		var slackUser SlackUser
		if err := json.NewDecoder(res.Body).Decode(&slackUser); err != nil {
			return nil, err
		}

		if !slackUser.Ok {
			return nil, errors.New("got unexpected response ok == false while contacting the Slack API")
		}

		return []string{
			fmt.Sprintf("slack:%s", slackUser.HTTPSSlackComTeamDomain),
		}, nil

	default:
		return nil, fmt.Errorf("implementation error: unsupported provider %#v", user.Provider)

	}
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

type SlackUser struct {
	Ok                        bool   `json:"ok"`
	Sub                       string `json:"sub"`
	HTTPSSlackComUserID       string `json:"https://slack.com/user_id"`
	HTTPSSlackComTeamID       string `json:"https://slack.com/team_id"`
	Email                     string `json:"email"`
	EmailVerified             bool   `json:"email_verified"`
	DateEmailVerified         int    `json:"date_email_verified"`
	Name                      string `json:"name"`
	Picture                   string `json:"picture"`
	GivenName                 string `json:"given_name"`
	FamilyName                string `json:"family_name"`
	Locale                    string `json:"locale"`
	HTTPSSlackComTeamName     string `json:"https://slack.com/team_name"`
	HTTPSSlackComTeamDomain   string `json:"https://slack.com/team_domain"`
	HTTPSSlackComUserImage24  string `json:"https://slack.com/user_image_24"`
	HTTPSSlackComUserImage32  string `json:"https://slack.com/user_image_32"`
	HTTPSSlackComUserImage48  string `json:"https://slack.com/user_image_48"`
	HTTPSSlackComUserImage72  string `json:"https://slack.com/user_image_72"`
	HTTPSSlackComUserImage192 string `json:"https://slack.com/user_image_192"`
	HTTPSSlackComUserImage512 string `json:"https://slack.com/user_image_512"`
	HTTPSSlackComTeamImage34  string `json:"https://slack.com/team_image_34"`
	HTTPSSlackComTeamImage44  string `json:"https://slack.com/team_image_44"`
	HTTPSSlackComTeamImage68  string `json:"https://slack.com/team_image_68"`
	HTTPSSlackComTeamImage88  string `json:"https://slack.com/team_image_88"`
	HTTPSSlackComTeamImage102 string `json:"https://slack.com/team_image_102"`
	HTTPSSlackComTeamImage132 string `json:"https://slack.com/team_image_132"`
	HTTPSSlackComTeamImage230 string `json:"https://slack.com/team_image_230"`
}

var forbiddenTableNames = map[string]struct{}{
	"create": {},
	"new":    {},
	"rand":   {},
	"random": {},
	"start":  {},
	"table":  {},
}

func normalizeTableName(input string) (scope string, id string, err error) {
	split := strings.Split(input, "/")

	switch len(split) {
	case 1:
		id = split[0]
	case 2:
		scope, id = split[0], split[1]
	default:
		err = fmt.Errorf("invalid table name %#v: too many slashes", input)
		return
	}

	id = strings.ToLower(id)
	id = strings.TrimSpace(id)
	// TODO: double slash
	// TODO: double space
	// TODO: trailing slash

	if _, ok := forbiddenTableNames[id]; ok {
		id = uuid.New().String()
		return
	}

	if len(id) == 0 {
		err = fmt.Errorf("invalid table name %#v: empty", input)
		return
	}

	return
}

package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const authBaseURL = "https://accounts.spotify.com/authorize"

type OAuthClient struct {
	ClientID    string
	RedirectURI string
	Scopes      []string
	HTTPClient  *http.Client
	Now         func() time.Time
}

func (c OAuthClient) AuthorizationURL(state, codeChallenge string) string {
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", c.ClientID)
	query.Set("redirect_uri", c.RedirectURI)
	query.Set("scope", strings.Join(c.Scopes, " "))
	query.Set("state", state)
	query.Set("code_challenge_method", "S256")
	query.Set("code_challenge", codeChallenge)

	return authBaseURL + "?" + query.Encode()
}

func (c OAuthClient) Exchange(ctx context.Context, code, codeVerifier string) (Token, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", c.RedirectURI)
	values.Set("client_id", c.ClientID)
	values.Set("code_verifier", codeVerifier)

	return requestToken(ctx, c.HTTPClient, c.Now, values)
}

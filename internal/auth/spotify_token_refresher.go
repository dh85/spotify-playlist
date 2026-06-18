package auth

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type SpotifyTokenRefresher struct {
	ClientID   string
	HTTPClient *http.Client
	Now        func() time.Time
}

func (r SpotifyTokenRefresher) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	values.Set("client_id", r.ClientID)

	return requestToken(ctx, r.HTTPClient, r.Now, values)
}

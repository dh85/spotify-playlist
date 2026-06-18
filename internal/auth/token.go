package auth

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	Expiry       time.Time `json:"expiry"`
}

func (t Token) IsValid(now time.Time) bool {
	if t.AccessToken == "" {
		return false
	}

	// Refresh a little early to avoid failing halfway through an API call.
	return now.Before(t.Expiry.Add(-60 * time.Second))
}

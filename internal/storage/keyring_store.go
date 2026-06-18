package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dh85/spotify-playlist/internal/auth"
	"github.com/zalando/go-keyring"
)

type KeyringTokenStore struct {
	Service string
	User    string
}

func (s KeyringTokenStore) Load(ctx context.Context) (auth.Token, error) {
	raw, err := keyring.Get(s.Service, s.User)
	if errors.Is(err, keyring.ErrNotFound) {
		return auth.Token{}, auth.ErrTokenNotFound
	}
	if err != nil {
		return auth.Token{}, err
	}

	var token auth.Token
	if err := json.Unmarshal([]byte(raw), &token); err != nil {
		return auth.Token{}, err
	}

	return token, nil
}

func (s KeyringTokenStore) Save(ctx context.Context, token auth.Token) error {
	data, _ := json.Marshal(token)
	return keyring.Set(s.Service, s.User, string(data))
}

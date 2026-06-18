package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrLoginRequired = errors.New("login required")
)

type TokenStore interface {
	Load(ctx context.Context) (Token, error)
	Save(ctx context.Context, token Token) error
}

type TokenRefresher interface {
	Refresh(ctx context.Context, refreshToken string) (Token, error)
}

type TokenManager struct {
	store     TokenStore
	refresher TokenRefresher
	now       func() time.Time

	mu sync.Mutex
}

func NewTokenManager(
	store TokenStore,
	refresher TokenRefresher,
	now func() time.Time,
) *TokenManager {
	return &TokenManager{
		store:     store,
		refresher: refresher,
		now:       now,
	}
}

func (m *TokenManager) ValidToken(ctx context.Context) (Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, err := m.store.Load(ctx)
	if errors.Is(err, ErrTokenNotFound) {
		return Token{}, ErrLoginRequired
	}
	if err != nil {
		return Token{}, err
	}

	if token.IsValid(m.now()) {
		return token, nil
	}

	if token.RefreshToken == "" {
		return Token{}, ErrLoginRequired
	}

	refreshed, err := m.refresher.Refresh(ctx, token.RefreshToken)
	if err != nil {
		return Token{}, err
	}

	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = token.RefreshToken
	}

	if err := m.store.Save(ctx, refreshed); err != nil {
		return Token{}, err
	}

	return refreshed, nil
}

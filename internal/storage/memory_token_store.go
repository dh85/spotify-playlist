package storage

import (
	"context"
	"sync"

	"github.com/dh85/spotify-playlist/internal/auth"
)

type MemoryTokenStore struct {
	mu    sync.RWMutex
	token *auth.Token
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

func (s *MemoryTokenStore) Load(ctx context.Context) (auth.Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == nil {
		return auth.Token{}, auth.ErrTokenNotFound
	}

	return *s.token, nil
}

func (s *MemoryTokenStore) Save(ctx context.Context, token auth.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = &token

	return nil
}

package auth

import (
	"context"
	"sync"
)

type SpyTokenStore struct {
	mu sync.Mutex

	token   Token
	saved   Token
	loadErr error
	saveErr error
}

func (s *SpyTokenStore) Load(ctx context.Context) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loadErr != nil {
		return Token{}, s.loadErr
	}

	return s.token, nil
}

func (s *SpyTokenStore) Save(ctx context.Context, token Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.saveErr != nil {
		return s.saveErr
	}

	s.saved = token
	s.token = token
	return nil
}

type SpyTokenRefresher struct {
	mu sync.Mutex

	token Token
	err   error
	calls int
}

func (s *SpyTokenRefresher) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls++

	if s.err != nil {
		return Token{}, s.err
	}

	return s.token, nil
}

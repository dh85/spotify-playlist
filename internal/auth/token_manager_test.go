package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestValidTokenReturnsStoredTokenWhenNotExpired(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(time.Hour),
		},
	}

	refresher := &SpyTokenRefresher{}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	got, err := manager.ValidToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got.AccessToken != "access-token" {
		t.Fatalf("access token = %q", got.AccessToken)
	}

	if refresher.calls != 0 {
		t.Fatalf("refresh calls = %d, want 0", refresher.calls)
	}
}

func TestValidTokenRefreshesExpiredTokenAndSavesIt(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "old-access-token",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(-time.Minute),
		},
	}

	refresher := &SpyTokenRefresher{
		token: Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			Expiry:       now.Add(time.Hour),
		},
	}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	got, err := manager.ValidToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got.AccessToken != "new-access-token" {
		t.Fatalf("access token = %q", got.AccessToken)
	}

	if refresher.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", refresher.calls)
	}

	if store.saved.AccessToken != "new-access-token" {
		t.Fatalf("saved token = %+v", store.saved)
	}
}

func TestValidTokenKeepsExistingRefreshTokenWhenSpotifyDoesNotReturnANewOne(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "old-access-token",
			RefreshToken: "existing-refresh-token",
			Expiry:       now.Add(-time.Minute),
		},
	}

	refresher := &SpyTokenRefresher{
		token: Token{
			AccessToken:  "new-access-token",
			RefreshToken: "",
			Expiry:       now.Add(time.Hour),
		},
	}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	got, err := manager.ValidToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got.RefreshToken != "existing-refresh-token" {
		t.Fatalf("refresh token = %q", got.RefreshToken)
	}
}

func TestValidTokenReturnsLoginRequiredWhenNoTokenIsStored(t *testing.T) {
	store := &SpyTokenStore{loadErr: ErrTokenNotFound}
	manager := NewTokenManager(store, &SpyTokenRefresher{}, time.Now)

	_, err := manager.ValidToken(context.Background())

	if !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("err = %v, want %v", err, ErrLoginRequired)
	}
}

func TestValidTokenReturnsErrorWhenLoadFails(t *testing.T) {
	loadErr := errors.New("disk failure")
	store := &SpyTokenStore{loadErr: loadErr}
	manager := NewTokenManager(store, &SpyTokenRefresher{}, time.Now)

	_, err := manager.ValidToken(context.Background())

	if err != loadErr {
		t.Fatalf("err = %v, want %v", err, loadErr)
	}
}

func TestValidTokenReturnsLoginRequiredWhenExpiredAndNoRefreshToken(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "expired",
			RefreshToken: "",
			Expiry:       now.Add(-time.Minute),
		},
	}

	manager := NewTokenManager(store, &SpyTokenRefresher{}, func() time.Time { return now })

	_, err := manager.ValidToken(context.Background())

	if !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("err = %v, want %v", err, ErrLoginRequired)
	}
}

func TestValidTokenReturnsErrorWhenRefreshFails(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "expired",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(-time.Minute),
		},
	}

	refreshErr := errors.New("network error")
	refresher := &SpyTokenRefresher{err: refreshErr}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	_, err := manager.ValidToken(context.Background())

	if err != refreshErr {
		t.Fatalf("err = %v, want %v", err, refreshErr)
	}
}

func TestValidTokenReturnsErrorWhenSaveFails(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	saveErr := errors.New("write failure")
	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "expired",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(-time.Minute),
		},
		saveErr: saveErr,
	}

	refresher := &SpyTokenRefresher{
		token: Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			Expiry:       now.Add(time.Hour),
		},
	}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	_, err := manager.ValidToken(context.Background())

	if err != saveErr {
		t.Fatalf("err = %v, want %v", err, saveErr)
	}
}

func TestValidTokenOnlyRefreshesOnceWhenCalledConcurrently(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	store := &SpyTokenStore{
		token: Token{
			AccessToken:  "old-access-token",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(-time.Minute),
		},
	}

	refresher := &SpyTokenRefresher{
		token: Token{
			AccessToken:  "new-access-token",
			RefreshToken: "refresh-token",
			Expiry:       now.Add(time.Hour),
		},
	}

	manager := NewTokenManager(store, refresher, func() time.Time { return now })

	const callers = 20

	var wg sync.WaitGroup
	wg.Add(callers)

	errs := make(chan error, callers)

	for range callers {
		go func() {
			defer wg.Done()
			_, err := manager.ValidToken(context.Background())
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	if refresher.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", refresher.calls)
	}
}

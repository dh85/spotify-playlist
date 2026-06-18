package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/dh85/spotify-playlist/internal/auth"
	"github.com/zalando/go-keyring"
)

func newTestStore() KeyringTokenStore {
	return KeyringTokenStore{Service: "test-service", User: "test-user"}
}

func TestKeyringTokenStore(t *testing.T) {
	ctx := context.Background()

	t.Run("Load returns ErrTokenNotFound when empty", func(t *testing.T) {
		keyring.MockInit()
		store := newTestStore()

		_, err := store.Load(ctx)
		if !errors.Is(err, auth.ErrTokenNotFound) {
			t.Fatalf("err = %v, want %v", err, auth.ErrTokenNotFound)
		}
	})

	t.Run("Save then Load returns the token", func(t *testing.T) {
		keyring.MockInit()
		store := newTestStore()

		token := auth.Token{AccessToken: "abc", RefreshToken: "def"}
		if err := store.Save(ctx, token); err != nil {
			t.Fatal(err)
		}

		got, err := store.Load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.AccessToken != "abc" || got.RefreshToken != "def" {
			t.Fatalf("got = %+v", got)
		}
	})

	t.Run("Load returns error on invalid JSON", func(t *testing.T) {
		keyring.MockInit()
		store := newTestStore()

		keyring.Set(store.Service, store.User, "not-json")

		_, err := store.Load(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Load returns error on keyring failure", func(t *testing.T) {
		keyring.MockInitWithError(errors.New("keyring locked"))
		store := newTestStore()

		_, err := store.Load(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Save returns error on keyring failure", func(t *testing.T) {
		keyring.MockInitWithError(errors.New("keyring locked"))
		store := newTestStore()

		err := store.Save(ctx, auth.Token{AccessToken: "x"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

}

package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/dh85/spotify-playlist/internal/auth"
)

func requireTokenNotFound(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, auth.ErrTokenNotFound) {
		t.Fatalf("err = %v, want %v", err, auth.ErrTokenNotFound)
	}
}

func TestMemoryTokenStore(t *testing.T) {
	ctx := context.Background()

	t.Run("Load returns ErrTokenNotFound when empty", func(t *testing.T) {
		store := NewMemoryTokenStore()

		_, err := store.Load(ctx)
		requireTokenNotFound(t, err)
	})

	t.Run("Save then Load returns the token", func(t *testing.T) {
		store := NewMemoryTokenStore()
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

}

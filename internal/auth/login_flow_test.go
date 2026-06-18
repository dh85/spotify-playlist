package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type spyBrowserOpener struct {
	url string
	err error
}

func (s *spyBrowserOpener) Open(u string) error {
	s.url = u
	return s.err
}

func (s *spyBrowserOpener) stateFromURL() string {
	parsed, _ := url.Parse(s.url)
	return parsed.Query().Get("state")
}

func (s *spyBrowserOpener) redirectWithCode(code string) func() (string, error) {
	return func() (string, error) {
		return "https://google.com/?code=" + code + "&state=" + s.stateFromURL(), nil
	}
}

func (s *spyBrowserOpener) redirectWithState(state string) func() (string, error) {
	return func() (string, error) {
		return "https://google.com/?state=" + state, nil
	}
}

func newTestFlow(browser *spyBrowserOpener, store *SpyTokenStore, redirectReader func() (string, error)) LoginFlow {
	return LoginFlow{
		OAuthClient: OAuthClient{
			ClientID:    "my-client",
			RedirectURI: "https://google.com/",
			Scopes:      []string{"user-read-private"},
		},
		Store:          store,
		BrowserOpener:  browser.Open,
		RedirectReader: redirectReader,
	}
}

func TestLoginFlow(t *testing.T) {
	fixedNow := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	t.Run("completes login flow successfully", func(t *testing.T) {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "new-access",
				"token_type": "Bearer",
				"scope": "user-read-private",
				"expires_in": 3600,
				"refresh_token": "new-refresh"
			}`))
		}))
		defer tokenServer.Close()

		store := &SpyTokenStore{}
		browser := &spyBrowserOpener{}

		flow := newTestFlow(browser, store, nil)
		flow.OAuthClient.HTTPClient = &http.Client{Transport: rewriteTo(tokenServer)}
		flow.OAuthClient.Now = func() time.Time { return fixedNow }
		flow.RedirectReader = browser.redirectWithCode("my-code")

		err := flow.Run(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if browser.url == "" {
			t.Error("expected browser to be opened")
		}
		if store.saved.AccessToken != "new-access" {
			t.Errorf("saved token = %+v", store.saved)
		}
	})

	t.Run("returns error when browser fails to open", func(t *testing.T) {
		browser := &spyBrowserOpener{err: errors.New("no browser")}
		flow := newTestFlow(browser, &SpyTokenStore{}, func() (string, error) { return "", nil })

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when redirect reader fails", func(t *testing.T) {
		browser := &spyBrowserOpener{}
		flow := newTestFlow(browser, &SpyTokenStore{}, func() (string, error) {
			return "", errors.New("user cancelled")
		})

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on state mismatch", func(t *testing.T) {
		browser := &spyBrowserOpener{}
		flow := newTestFlow(browser, &SpyTokenStore{}, func() (string, error) {
			return "https://google.com/?code=x&state=wrong", nil
		})

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when Spotify returns error param", func(t *testing.T) {
		browser := &spyBrowserOpener{}
		flow := newTestFlow(browser, &SpyTokenStore{}, func() (string, error) {
			return "https://google.com/?error=access_denied&state=whatever", nil
		})

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when redirect URL has no code", func(t *testing.T) {
		browser := &spyBrowserOpener{}
		flow := newTestFlow(browser, &SpyTokenStore{}, browser.redirectWithState(""))

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when token exchange fails", func(t *testing.T) {
		tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer tokenServer.Close()

		browser := &spyBrowserOpener{}
		flow := newTestFlow(browser, &SpyTokenStore{}, nil)
		flow.OAuthClient.HTTPClient = &http.Client{Transport: rewriteTo(tokenServer)}
		flow.OAuthClient.Now = func() time.Time { return fixedNow }
		flow.RedirectReader = browser.redirectWithCode("my-code")

		err := flow.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

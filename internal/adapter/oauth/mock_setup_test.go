package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newMockRedis creates a new miniredis server and returns the redis client and cleanup function
func newMockRedis() (*redis.Client, func(), error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup, nil
}

// mockOAuthServerConfig configures the mock OAuth server responses.
type mockOAuthServerConfig struct {
	TokenResponse  map[string]any
	TokenStatus    int
	UserInfoResp   *googleUserResp
	UserInfoStatus int
}

// setupMockHTTPServer creates a mock HTTP server for OAuth endpoints.
// Returns server URL and cleanup function.
func setupMockHTTPServer(t *testing.T, cfg *mockOAuthServerConfig) (*httptest.Server, func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/token"):
			// Token exchange endpoint
			w.Header().Set("Content-Type", "application/json")
			if cfg.TokenStatus != 0 {
				w.WriteHeader(cfg.TokenStatus)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			if cfg.TokenResponse != nil {
				json.NewEncoder(w).Encode(cfg.TokenResponse)
			}

		case strings.HasPrefix(r.URL.Path, "/oauth2/v2/userinfo"):
			// User info endpoint
			w.Header().Set("Content-Type", "application/json")
			if cfg.UserInfoStatus != 0 {
				w.WriteHeader(cfg.UserInfoStatus)
			} else {
				w.WriteHeader(http.StatusOK)
			}

			if cfg.UserInfoResp != nil {
				json.NewEncoder(w).Encode(cfg.UserInfoResp)
			} else if cfg.UserInfoStatus == http.StatusOK {
				// Return malformed JSON for invalid response test
				w.Write([]byte("invalid json"))
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

// mockTransport is a custom http.RoundTripper that redirects requests to the mock server.
type mockTransport struct {
	serverURL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new request pointing to the mock server
	mockReq := req.Clone(req.Context())
	mockReq.URL, _ = url.Parse(m.serverURL)

	// Use the default transport to make the request
	return http.DefaultTransport.RoundTrip(mockReq)
}

// failingTransport is a custom http.RoundTripper that always returns an error.
type failingTransport struct {
	err error
}

func (f *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("failed to get user info: %w", f.err)
}

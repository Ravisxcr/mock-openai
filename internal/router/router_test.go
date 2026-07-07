package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestHealthCheckDoesNotRequireAuth(t *testing.T) {
	r := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"running"}`, w.Body.String())
}

func TestV1RoutesRequireAuth(t *testing.T) {
	r := SetupRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"chat completions", http.MethodPost, "/v1/chat/completions"},
		{"list chat completions", http.MethodGet, "/v1/chat/completions"},
		{"models", http.MethodGet, "/v1/models"},
		{"embeddings", http.MethodPost, "/v1/embeddings"},
		{"batches", http.MethodGet, "/v1/batches"},
		{"files", http.MethodGet, "/v1/files"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	r := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/models", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "You didn't provide an API key")
}

func TestAuthMiddlewareMalformedHeader(t *testing.T) {
	r := SetupRouter()

	cases := []string{
		"Basic sometoken", // wrong scheme
		"Bearer",          // missing token
		"Bearer ",         // empty token
		"justatoken",      // no scheme at all
	}

	for _, header := range cases {
		t.Run(header, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/models", nil)
			req.Header.Set("Authorization", header)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddlewareValidBearerToken(t *testing.T) {
	r := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer sk-test-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotFoundRoute(t *testing.T) {
	r := SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/v1/does-not-exist", nil)
	req.Header.Set("Authorization", "Bearer sk-test-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

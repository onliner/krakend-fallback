package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newNext(status int, contentType string, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func doRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "http://example.com"+path, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func decodeJSON(t *testing.T, b []byte) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	assert.NoError(t, err)

	return m
}

func TestHandler_RouteNotMatched_PassesThrough(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default:  map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(200, "text/plain", "ok")
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/health")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestHandler_Not2xx_PassesThrough(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{Path: "/products/{product}/positions"},
		},
	}

	next := newNext(500, "application/json", `{"error":"boom"}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"boom"}`, rr.Body.String())
}

func TestHandler_NotJSON_PassesThrough(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{Path: "/products/{product}/positions"},
		},
	}

	next := newNext(200, "text/plain; charset=utf-8", "hello")
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "hello", rr.Body.String())
}

func TestHandler_JSONSuccess_AddsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default: map[string]interface{}{
					"positions": []interface{}{},
					"shops":     nil,
				},
			},
		},
	}

	next := newNext(200, "application/json; charset=utf-8", `{"product":{"id":1}}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())

	_, ok := out["product"]
	assert.True(t, ok)
	_, ok = out["positions"]
	assert.True(t, ok)
	_, ok = out["shops"]
	assert.True(t, ok)
}

func TestHandler_JSONSuccess_MissingRequired_ReturnsServerError(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default:  map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(200, "application/json", `{"positions":[]}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())

	assert.Equal(t, http.StatusText(http.StatusInternalServerError), out["message"])
}

func TestHandler_InvalidJSON_ReturnsServerError(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default:  map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(200, "application/json", `{"product":`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), out["message"])
}

func TestHandler_MissingRequired_ReturnsFirstBackendError(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default:  map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(200, "application/json", `{"positions":[],"error_2":{"http_status_code":502,"http_body":"second","http_body_encoding":"text/plain"},"error_1":{"http_status_code":500,"http_body":"first","http_body_encoding":"text/plain"}}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "first", rr.Body.String())
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
}

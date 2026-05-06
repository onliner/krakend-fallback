package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestHandler_Not2xx_JSON_ReturnsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:    "/products/{product}/positions",
				Default: map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(500, "application/json", `{"error":"boom"}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	assert.Equal(t, []interface{}{}, out["positions"])
	_, hasError := out["error"]
	assert.False(t, hasError)
}

func TestHandler_401_JSON_ReturnsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:    "/products/{product}/positions",
				Default: map[string]interface{}{"positions": []interface{}{}, "shops": nil},
			},
		},
	}

	next := newNext(401, "application/json", `{"error":"unauthorized"}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	_, ok := out["positions"]
	assert.True(t, ok)
	_, ok = out["shops"]
	assert.True(t, ok)
	_, hasError := out["error"]
	assert.False(t, hasError)
}

func TestHandler_403_JSON_ReturnsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:    "/products/{product}/positions",
				Default: map[string]interface{}{"positions": []interface{}{}, "extra": "default_val"},
			},
		},
	}

	next := newNext(403, "application/json", `{"error":"forbidden","positions":[1,2]}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	assert.Equal(t, []interface{}{}, out["positions"])
	assert.Equal(t, "default_val", out["extra"])
}

func TestHandler_Not2xx_NotJSON_ReturnsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:    "/products/{product}/positions",
				Default: map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(500, "text/plain", "internal error")
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	assert.Equal(t, []interface{}{}, out["positions"])
}

func TestHandler_NotJSON_ReturnsDefaults(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:    "/products/{product}/positions",
				Default: map[string]interface{}{"positions": []interface{}{}},
			},
		},
	}

	next := newNext(200, "text/plain; charset=utf-8", "hello")
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())
	assert.Equal(t, []interface{}{}, out["positions"])
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

func TestHandler_JSONSuccess_PreservesExistingValues(t *testing.T) {
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

	next := newNext(200, "application/json; charset=utf-8", `{"product":{"id":1},"positions":[1,2,3]}`)
	h := NewHandler(cfg, next, noopLogger{})

	rr := doRequest(t, h, "GET", "/products/iphonex64s/positions")

	assert.Equal(t, http.StatusOK, rr.Code)

	out := decodeJSON(t, rr.Body.Bytes())

	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, out["positions"])
	_, ok := out["shops"]
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

func TestHandler_ServeHTTP_ConcurrentAccess(t *testing.T) {
	t.Parallel()

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

	const workers = 64

	var wg sync.WaitGroup
	errs := make(chan string, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "http://example.com/products/iphonex64s/positions", nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				errs <- "unexpected status code"
				return
			}

			var out map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
				errs <- "invalid json response"
				return
			}

			if _, ok := out["product"]; !ok {
				errs <- "missing required key: product"
				return
			}
			if _, ok := out["positions"]; !ok {
				errs <- "missing default key: positions"
				return
			}
			if _, ok := out["shops"]; !ok {
				errs <- "missing default key: shops"
				return
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		assert.Fail(t, err)
	}
}

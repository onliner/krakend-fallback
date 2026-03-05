package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBackendError_NoKeys(t *testing.T) {
	resp, ok := FindBackendError(map[string]interface{}{
		"product": map[string]interface{}{"id": 1},
	})

	assert.False(t, ok)
	assert.Nil(t, resp)
}

func TestFindBackendError_PicksFirstSortedKey(t *testing.T) {
	body := map[string]interface{}{
		"error_2": map[string]interface{}{
			"http_status_code":   502,
			"http_body":          `{"message":"second"}`,
			"http_body_encoding": "application/json",
		},
		"error_1": map[string]interface{}{
			"http_status_code":   500,
			"http_body":          `{"message":"first"}`,
			"http_body_encoding": "application/json; charset=utf-8",
		},
		"product": map[string]interface{}{"id": 1},
	}

	resp, ok := FindBackendError(body)
	assert.True(t, ok)
	assert.NotNil(t, resp)

	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	gotBody := readString(t, resp.Body)
	assert.Equal(t, `{"message":"first"}`, gotBody)
	assert.Equal(t, int64(len(gotBody)), resp.ContentLength)
}

func TestFindBackendError_InvalidShape_ReturnsFalse(t *testing.T) {
	body := map[string]interface{}{
		"error_1": "not an object",
	}

	resp, ok := FindBackendError(body)
	assert.False(t, ok)
	assert.Nil(t, resp)
}

func TestFindBackendErrorMissingEncoding(t *testing.T) {
	body := map[string]interface{}{
		"error_1": map[string]interface{}{
			"http_status_code": 503,
			"http_body":        "service unavailable",
		},
	}

	resp, ok := FindBackendError(body)
	assert.True(t, ok)
	assert.NotNil(t, resp)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))

	got := readString(t, resp.Body)
	assert.Equal(t, `service unavailable`, got)
}

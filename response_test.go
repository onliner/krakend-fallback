package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSuccess(t *testing.T) {
	cases := []struct {
		code int
		want bool
	}{
		{199, false},
		{200, true},
		{204, true},
		{299, true},
		{300, false},
		{404, false},
		{500, false},
	}

	for _, tt := range cases {
		resp := &http.Response{StatusCode: tt.code}
		assert.Equal(t, tt.want, IsSuccess(resp))
	}
}

func TestNewJsonResponse_SetsFieldsBodyAndPreservesHeaders(t *testing.T) {
	inHeaders := http.Header{
		"X-Test":         []string{"a", "b"},
		"Cache-Control":  []string{"no-cache"},
		"Content-Type":   []string{"text/plain; charset=utf-8"},
		"Content-Length": []string{"999"},
	}

	body := map[string]interface{}{
		"foo": "bar",
		"n":   float64(1),
	}

	resp := NewJsonResponse(http.StatusCreated, inHeaders, body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "", resp.Header.Get("Content-Length"))
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))

	b := readBytes(t, resp.Body)

	assert.Equal(t, int64(len(b)), resp.ContentLength)

	var got map[string]interface{}
	err := json.Unmarshal(b, &got)
	assert.NoError(t, err)

	assert.Equal(t, "bar", got["foo"])
	assert.Equal(t, float64(1), got["n"])
}

func TestContentType(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=utf-8"},
		},
	}
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestIsJSON(t *testing.T) {
	cases := []struct {
		ct   string
		want bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"APPLICATION/JSON", true},
		{"application/problem+json", true},
		{"application/vnd.api+json", true},
		{"text/plain", false},
		{"", false},
	}

	for _, tt := range cases {
		resp := &http.Response{
			Header: http.Header{"Content-Type": []string{tt.ct}},
		}
		assert.Equal(t, tt.want, IsJSON(resp))
	}
}

func TestNewServerError(t *testing.T) {
	resp := NewServerError()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	b := readBytes(t, resp.Body)

	var got map[string]interface{}
	err := json.Unmarshal(b, &got)
	assert.NoError(t, err)

	msg, _ := got["message"].(string)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), msg)
}

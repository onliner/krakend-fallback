package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readBody(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		assert.NoError(t, err)
	}(rc)
	b, err := io.ReadAll(rc)
	assert.NoError(t, err)

	return string(b)
}

func TestResponseRecorder_DefaultStatusIs200(t *testing.T) {
	rr := NewResponseRecorder()
	resp := rr.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Zero(t, resp.ContentLength)
	assert.NotNil(t, resp.Body)

	body := readBody(t, resp.Body)
	assert.Empty(t, body)
}

func TestResponseRecorder_WriteHeaderFirstWins(t *testing.T) {
	rr := NewResponseRecorder()

	rr.WriteHeader(http.StatusCreated)
	rr.WriteHeader(http.StatusTeapot)

	resp := rr.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestResponseRecorder_HeadersRecorded(t *testing.T) {
	rr := NewResponseRecorder()
	rr.Header().Add("X-Test", "a")
	rr.Header().Add("X-Test", "b")
	rr.Header().Set("Content-Type", "application/json")

	resp := rr.Result()

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestResponseRecorder_BodyAndContentLength(t *testing.T) {
	rr := NewResponseRecorder()

	_, err := rr.Write([]byte("he"))
	assert.NoError(t, err)
	_, err = rr.Write([]byte("llo"))
	assert.NoError(t, err)

	resp := rr.Result()
	assert.Equal(t, int64(len("hello")), resp.ContentLength)

	body := readBody(t, resp.Body)
	assert.Equal(t, "hello", body)
}

func TestResponseRecorder_ResultBodyStartsFromBeginning(t *testing.T) {
	rr := NewResponseRecorder()
	_, _ = rr.Write([]byte(`{"a":1}`))

	resp := rr.Result()
	body := readBody(t, resp.Body)

	assert.Equal(t, `{"a":1}`, body)
}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func IsSuccess(r *http.Response) bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

func NewJsonResponse(status int, headers http.Header, body map[string]interface{}) *http.Response {
	hdr := headers.Clone()
	hdr.Del("Content-Length")
	hdr.Set("Content-Type", "application/json")
	out, _ := json.Marshal(body)

	return &http.Response{
		StatusCode:    status,
		Header:        hdr,
		Body:          io.NopCloser(bytes.NewBuffer(out)),
		ContentLength: int64(len(out)),
	}
}

func ContentType(r *http.Response) string {
	return r.Header.Get("Content-Type")
}

func IsJSON(r *http.Response) bool {
	ct := strings.ToLower(ContentType(r))
	return strings.Contains(ct, "application/json") || strings.Contains(ct, "+json")
}

func NewServerError() *http.Response {
	status := http.StatusInternalServerError

	return NewJsonResponse(status, http.Header{}, map[string]interface{}{"message": http.StatusText(status)})
}

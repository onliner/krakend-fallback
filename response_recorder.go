package main

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseRecorder struct {
	hdr         http.Header
	body        *bytes.Buffer
	status      int
	wroteHeader bool
}

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		hdr:    make(http.Header),
		body:   &bytes.Buffer{},
		status: http.StatusOK,
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.hdr
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.status = statusCode
	r.wroteHeader = true
}

func (r *ResponseRecorder) Write(p []byte) (int, error) {
	return r.body.Write(p)
}

func (r *ResponseRecorder) Result() *http.Response {
	code := r.status

	return &http.Response{
		StatusCode:    code,
		Header:        r.hdr,
		Body:          io.NopCloser(bytes.NewReader(r.body.Bytes())),
		ContentLength: int64(r.body.Len()),
	}
}

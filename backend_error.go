package main

import (
	"bytes"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type BackendError struct {
	Status   int    `mapstructure:"http_status_code"`
	Body     string `mapstructure:"http_body"`
	Encoding string `mapstructure:"http_body_encoding"`
}

func (e BackendError) ToResponse() *http.Response {
	enc := e.Encoding
	if enc == "" {
		enc = "text/plain"
	}
	return &http.Response{
		StatusCode:    e.Status,
		Header:        http.Header{"Content-Type": []string{enc}},
		Body:          io.NopCloser(bytes.NewBufferString(e.Body)),
		ContentLength: int64(len(e.Body)),
	}
}

func FindBackendError(body map[string]interface{}) (*http.Response, bool) {
	keys := make([]string, 0, len(body))
	for k := range body {
		if strings.HasPrefix(k, "error_") {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, false
	}

	sort.Strings(keys)
	v, ok := body[keys[0]]
	if !ok {
		return nil, false
	}

	var berr BackendError

	if err := mapstructure.WeakDecode(v, &berr); err != nil {
		return nil, false
	}

	return berr.ToResponse(), true
}

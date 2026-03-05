package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Handler struct {
	config *Config
	next   http.Handler
	logger Logger
}

func NewHandler(config *Config, next http.Handler, logger Logger) *Handler {
	return &Handler{config, next, logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, ok := h.config.MatchRoute(req.URL.Path)

	if !ok {
		h.next.ServeHTTP(w, req)
		return
	}

	rec := NewResponseRecorder()
	h.next.ServeHTTP(rec, req)
	resp := rec.Result()

	if !IsSuccess(resp) || !IsJSON(resp) {
		h.write(w, resp)
		return
	}

	var body map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		h.write(w, NewServerError())
		h.logger.Error(fmt.Errorf("error %s decoding response body: %v", req.URL.Path, err))
		return
	}

	for _, k := range route.Required {
		if _, exists := body[k]; !exists {
			h.logger.Debug(fmt.Sprintf("[PLUGIN: %s] required field %q missing in %s", Namespace, k, req.URL.Path))
			errResp, ok := FindBackendError(body)
			if !ok {
				errResp = NewServerError()
			}
			h.write(w, errResp)
			return
		}
	}

	for k, v := range route.Default {
		if _, exists := body[k]; !exists {
			body[k] = v
		}
	}

	h.write(w, NewJsonResponse(resp.StatusCode, resp.Header, body))
}

func (h *Handler) write(w http.ResponseWriter, res *http.Response) {
	for k, hs := range res.Header {
		w.Header().Del(k)
		for _, v := range hs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(res.StatusCode)

	if res.Body == nil {
		return
	}

	_, err := io.Copy(w, res.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed write response body: %v", err))
	}

	_ = res.Body.Close()
}

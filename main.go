package main

import (
	"context"
	"fmt"
	"net/http"
)

func main() {}

var HandlerRegisterer = registerer(Namespace)
var logger Logger = nil

type registerer string

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", Namespace))
}

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.wrap)
}

func (r registerer) wrap(_ context.Context, extra map[string]interface{}, next http.Handler) (http.Handler, error) {
	cfg, err := NewConfig(extra)
	if err != nil {
		return next, err
	}

	return NewHandler(cfg, next, logger), nil
}

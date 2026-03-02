package main

import (
	"errors"

	"github.com/mitchellh/mapstructure"
)

const Namespace = "onliner/krakend-fallback"

type Config struct {
	Routes []Route `mapstructure:"routes"`
}

type Route struct {
	Path     string                 `mapstructure:"path"`
	Required []string               `mapstructure:"required"`
	Default  map[string]interface{} `mapstructure:"default"`
}

func NewConfig(input map[string]interface{}) (*Config, error) {
	raw, ok := input[Namespace].(map[string]interface{})
	if !ok {
		return nil, errors.New("configuration not found")
	}

	var cfg Config

	if err := mapstructure.WeakDecode(raw, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) MatchRoute(path string) (Route, bool) {
	for _, r := range cfg.Routes {
		if MatchPathTemplate(r.Path, path) {
			return r, true
		}
	}

	return Route{}, false
}

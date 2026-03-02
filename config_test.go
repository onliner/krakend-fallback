package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_ConfigNotFound(t *testing.T) {
	_, err := NewConfig(map[string]interface{}{})
	assert.Error(t, errors.New("configuration not found"), err)
}

func TestNewConfig_DecodeOK(t *testing.T) {
	raw := []byte(`{
	  "onliner/krakend-fallback": {
	    "routes": [
	      {
	        "path": "/products/{product}/positions",
	        "required": ["product"],
	        "default": {"positions": [], "shops": null}
	      },
	      {
	        "path": "/health",
	        "required": [],
	        "default": {}
	      }
	    ]
	  }
	}`)

	var input map[string]interface{}
	err := json.Unmarshal(raw, &input)
	assert.NoError(t, err)

	cfg, err := NewConfig(input)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Routes))

	r0 := cfg.Routes[0]
	assert.Equal(t, "/products/{product}/positions", r0.Path)
	assert.Equal(t, len(r0.Required), 1)
	assert.Equal(t, "product", r0.Required[0])

	_, ok := r0.Default["positions"]
	assert.True(t, ok)
	_, ok = r0.Default["shops"]
	assert.True(t, ok)
}

func TestConfig_MatchRoute_MatchesTemplate(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{
				Path:     "/products/{product}/positions",
				Required: []string{"product"},
				Default:  map[string]interface{}{"positions": []interface{}{}},
			},
			{
				Path:     "/health",
				Required: nil,
				Default:  nil,
			},
		},
	}

	route, ok := cfg.MatchRoute("/products/iphonex64s/positions")
	assert.True(t, ok)
	assert.Equal(t, "/products/{product}/positions", route.Path)
}

func TestConfig_MatchRoute_NoMatch(t *testing.T) {
	cfg := &Config{
		Routes: []Route{
			{Path: "/a/{x}/b"},
		},
	}

	_, ok := cfg.MatchRoute("/a/1/c")
	assert.False(t, ok)
}

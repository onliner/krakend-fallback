package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPathTemplate(t *testing.T) {
	tests := []struct {
		name string
		tpl  string
		path string
		want bool
	}{
		{
			name: "matches single param",
			tpl:  "/products/{product}/positions",
			path: "/products/iphonex64s/positions",
			want: true,
		},
		{
			name: "matches with leading/trailing slashes",
			tpl:  "///products/{product}/positions///",
			path: "/products/iphonex64s/positions/",
			want: true,
		},
		{
			name: "does not match different static segment",
			tpl:  "/products/{product}/positions",
			path: "/items/iphonex64s/positions",
			want: false,
		},
		{
			name: "does not match different suffix segment",
			tpl:  "/products/{product}/positions",
			path: "/products/iphonex64s/price",
			want: false,
		},
		{
			name: "does not match when segment count differs (extra)",
			tpl:  "/products/{product}/positions",
			path: "/products/iphonex64s/positions/extra",
			want: false,
		},
		{
			name: "does not match when segment count differs (missing)",
			tpl:  "/products/{product}/positions",
			path: "/products/iphonex64s",
			want: false,
		},
		{
			name: "param must match non-empty segment",
			tpl:  "/products/{product}/positions",
			path: "/products//positions",
			want: false,
		},
		{
			name: "no params exact match",
			tpl:  "/a/b/c",
			path: "/a/b/c",
			want: true,
		},
		{
			name: "no params exact mismatch",
			tpl:  "/a/b/c",
			path: "/a/b/d",
			want: false,
		},
		{
			name: "root matches root",
			tpl:  "/",
			path: "/",
			want: true,
		},
		{
			name: "root does not match non-root",
			tpl:  "/",
			path: "/a",
			want: false,
		},
		{
			name: "treat braces-only segments as params",
			tpl:  "/x/{id}/y/{slug}",
			path: "/x/123/y/abc",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MatchPathTemplate(tt.tpl, tt.path))
		})
	}
}

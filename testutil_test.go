package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readBytes(t *testing.T, rc io.ReadCloser) []byte {
	t.Helper()
	if rc == nil {
		return nil
	}

	defer func() {
		err := rc.Close()
		assert.NoError(t, err)
	}()

	b, err := io.ReadAll(rc)
	assert.NoError(t, err)

	return b
}

func readString(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	return string(readBytes(t, rc))
}

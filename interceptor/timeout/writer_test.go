// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Copied from https://github.com/gin-contrib/timeout/blob/master/writer_test.go
package rkgintimeout

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteHeader(t *testing.T) {
	code1 := 99
	errmsg1 := fmt.Sprintf("invalid http status code: %d", code1)
	code2 := 1000
	errmsg2 := fmt.Sprintf("invalid http status code: %d", code2)

	writer := writer{}
	assert.PanicsWithValue(t, errmsg1, func() {
		writer.WriteHeader(code1)
	})
	assert.PanicsWithValue(t, errmsg2, func() {
		writer.WriteHeader(code2)
	})
}

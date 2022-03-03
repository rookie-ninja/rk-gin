// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Copied from https://github.com/gin-contrib/timeout/blob/master/buffer_pool_test.go
package rkgintout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuffer(t *testing.T) {
	pool := &bufferPool{}
	buf := pool.Get()
	assert.NotEqual(t, nil, buf)
	pool.Put(buf)
	buf2 := pool.Get()
	assert.NotEqual(t, nil, buf2)
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Copied from https://github.com/gin-contrib/timeout/blob/master/buffer_pool.go

package rkgintout

import (
	"bytes"
	"sync"
)

// bufferPool is Pool of *bytes.Buffer
type bufferPool struct {
	pool sync.Pool
}

// Get a bytes.Buffer pointer
func (p *bufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get()
	if buf == nil {
		return &bytes.Buffer{}
	}
	return buf.(*bytes.Buffer)
}

// Put a bytes.Buffer pointer to BufferPool
func (p *bufferPool) Put(buf *bytes.Buffer) {
	p.pool.Put(buf)
}

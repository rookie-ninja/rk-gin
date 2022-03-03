// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgingzip

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		data:   make([]byte, 0),
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	data       []byte
	statusCode int
	header     http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	m.data = bytes
	return len(bytes), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestNewOptionSet(t *testing.T) {
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())

	// without options
	set := newOptionSet()
	assert.NotEmpty(t, set.EntryName)
	assert.False(t, set.Skipper(ctx))
	assert.Equal(t, DefaultCompression, set.Level)
	assert.NotNil(t, set.decompressPool)
	assert.NotNil(t, set.compressPool)

	// with level
	set = newOptionSet(
		WithEntryNameAndType("ut-name", "ut-type"),
		WithLevel(NoCompression),
		WithSkipper(func(context *gin.Context) bool {
			return true
		}))
	assert.Equal(t, NoCompression, set.Level)
}

func TestNewCompressPool(t *testing.T) {
	// with DefaultCompression
	pool := newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with NoCompression
	pool = newCompressPool(NoCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(BestSpeed)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(BestCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(HuffmanOnly)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool("invalid")
	assert.NotNil(t, pool.delegate.Get())
}

func TestCompressPool_Get(t *testing.T) {
	pool := newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.Get())
}

func TestCompressPool_Put(t *testing.T) {
	defer assertNotPanic(t)

	pool := newCompressPool(DefaultCompression)
	// put different types of value
	pool.Put(nil)
	pool.Put("string")
	pool.Put(1)
}

func TestDecompressPool_Get(t *testing.T) {
	pool := newDecompressPool()
	assert.NotNil(t, pool.Get())
}

func TestDecompressPool_Put(t *testing.T) {
	defer assertNotPanic(t)

	pool := newDecompressPool()
	// put different types of value
	pool.Put(nil)
	pool.Put("string")
	pool.Put(1)
}

func TestGzipResponseWriter(t *testing.T) {
	defer assertNotPanic(t)

	// WriteHeader() write header with http.StatusNoContent
	rw := NewMockResponseWriter()
	ctx, _ := gin.CreateTestContext(rw)
	w := gzip.NewWriter(new(bytes.Buffer))
	gzipRW := newGzipResponseWriter(w, ctx.Writer)
	gzipRW.WriteHeader(http.StatusNoContent)
	assert.Empty(t, rw.Header().Get(headerContentEncoding))

	// WriteHeader() write header with other status code
	rw = NewMockResponseWriter()
	ctx, _ = gin.CreateTestContext(rw)
	w = gzip.NewWriter(new(bytes.Buffer))
	gzipRW = newGzipResponseWriter(w, ctx.Writer)
	gzipRW.WriteHeader(http.StatusOK)

	// Write() without Content-Type
	rw = NewMockResponseWriter()
	ctx, _ = gin.CreateTestContext(rw)
	buf := new(bytes.Buffer)
	w = gzip.NewWriter(buf)
	gzipRW = newGzipResponseWriter(w, ctx.Writer)
	gzipRW.Write([]byte("ut-message"))
	assert.Empty(t, rw.Header().Get(headerContentLength))
	assert.NotEmpty(t, buf.String())

	// Write() with Content-Type
	rw = NewMockResponseWriter()
	ctx, _ = gin.CreateTestContext(rw)
	buf = new(bytes.Buffer)
	w = gzip.NewWriter(buf)
	gzipRW = newGzipResponseWriter(w, ctx.Writer)
	rw.Header().Set(headerContentType, "ut-type")
	gzipRW.Write([]byte("ut-message"))
	assert.NotEmpty(t, rw.Header().Get(headerContentType))
	assert.NotEmpty(t, buf.String())
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

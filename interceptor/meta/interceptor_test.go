// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginmeta

import (
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

func TestInterceptor(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	handler(ctx)

	assert.NotEmpty(t, ctx.Writer.Header().Get("X-RK-App-Name"))
	assert.Empty(t, ctx.Writer.Header().Get("X-RK-App-Version"))
	assert.NotEmpty(t, ctx.Writer.Header().Get("X-RK-App-Unix-Time"))
	assert.NotEmpty(t, ctx.Writer.Header().Get("X-RK-Received-Time"))
}

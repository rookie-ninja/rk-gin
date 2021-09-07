// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintrace

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
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
	defer assertNotPanic(t)
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithExporter(&NoopExporter{}))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
	}

	ctx.Error(errors.New("ut error"))
	handler(ctx)
}

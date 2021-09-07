// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginauth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/stretchr/testify/assert"
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

func TestInterceptor_WithIgnoringPath(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"),
		WithIgnorePrefix("ut-ignore-path"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-ignore-path",
		},
	}

	handler(ctx)
	assert.False(t, ctx.IsAborted())
}

func TestInterceptor_WithBasicAuth_Invalid(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: http.Header{},
	}

	ctx.Request.Header.Set(rkgininter.RpcAuthorizationHeaderKey, "invalid")

	handler(ctx)
	assert.True(t, ctx.IsAborted())
}

func TestInterceptor_WithBasicAuth_InvalidBasicAuth(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: http.Header{},
	}

	ctx.Request.Header.Set(rkgininter.RpcAuthorizationHeaderKey, fmt.Sprintf("%s invalid", typeBasic))

	handler(ctx)
	assert.True(t, ctx.IsAborted())
}

func TestInterceptor_WithApiKey_Invalid(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: http.Header{},
	}

	ctx.Request.Header.Set(rkgininter.RpcApiKeyHeaderKey, "invalid")

	handler(ctx)
	assert.True(t, ctx.IsAborted())
}

func TestInterceptor_MissingAuth(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: http.Header{},
	}

	handler(ctx)
	assert.True(t, ctx.IsAborted())
}

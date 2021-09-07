// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgininter

import (
	"github.com/gin-gonic/gin"
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

func TestGetRemoteAddressSet(t *testing.T) {
	// With nil context
	ip, port := GetRemoteAddressSet(nil)
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With nil Request in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ip, port = GetRemoteAddressSet(ctx)
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With x-forwarded-for equals to ::1
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		RemoteAddr: "1.1.1.1:1",
		Header:     http.Header{},
	}
	ctx.Request.Header.Set("x-forwarded-for", "::1")
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "localhost", ip)
	assert.Equal(t, "1", port)

	// Happy case
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		RemoteAddr: "1.1.1.1:1",
		Header:     http.Header{},
	}
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "1.1.1.1", ip)
	assert.Equal(t, "1", port)
}

func TestShouldLog(t *testing.T) {
	// With nil context
	assert.False(t, ShouldLog(nil))

	// With nil request in context
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	assert.False(t, ShouldLog(ctx))

	// With ignoring path
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/assets",
		},
	}
	assert.False(t, ShouldLog(ctx))
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/tv",
		},
	}
	assert.False(t, ShouldLog(ctx))
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/sw/",
		},
	}
	assert.False(t, ShouldLog(ctx))

	// Expect true
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
	}
	assert.True(t, ShouldLog(ctx))
}

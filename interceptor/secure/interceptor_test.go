// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginsec

import (
	"crypto/tls"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func newCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		Header: make(map[string][]string, 0),
		URL:    &url.URL{},
	}

	return ctx
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

	// with skipper
	handler := Interceptor(WithSkipper(func(context *gin.Context) bool {
		return true
	}))
	ctx := newCtx()
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	// without options
	handler = Interceptor()
	ctx = newCtx()
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	containsHeader(t, ctx,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions)

	// with options
	handler = Interceptor(
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix"))
	ctx = newCtx()
	ctx.Request.TLS = &tls.ConnectionState{}
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	containsHeader(t, ctx,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions,
		headerStrictTransportSecurity,
		headerContentSecurityPolicyReportOnly,
		headerReferrerPolicy)
}

func containsHeader(t *testing.T, ctx *gin.Context, headers ...string) {
	for _, v := range headers {
		assert.Contains(t, ctx.Writer.Header(), v)
	}
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

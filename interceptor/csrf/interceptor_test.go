// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgincsrf

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// match 1
	handler := Interceptor(WithSkipper(func(context *gin.Context) bool {
		return true
	}))
	ctx := newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	// match 2.1
	handler = Interceptor()
	ctx = newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Contains(t, ctx.Writer.Header().Get("Set-Cookie"), "_csrf")

	// match 2.2
	handler = Interceptor()
	ctx = newCtx(http.MethodGet)
	ctx.Request.AddCookie(&http.Cookie{
		Name:  "_csrf",
		Value: "ut-csrf-token",
	})
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Contains(t, ctx.Writer.Header().Get("Set-Cookie"), "_csrf")

	// match 3.1
	handler = Interceptor()
	ctx = newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	// match 3.2
	handler = Interceptor()
	ctx = newCtx(http.MethodPost)
	handler(ctx)
	assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())

	// match 3.3
	handler = Interceptor()
	ctx = newCtx(http.MethodPost)
	ctx.Request.Header.Set(headerXCSRFToken, "ut-csrf-token")
	handler(ctx)
	assert.Equal(t, http.StatusForbidden, ctx.Writer.Status())

	// match 4.1
	handler = Interceptor(WithCookiePath("ut-path"))
	ctx = newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Contains(t, ctx.Writer.Header().Get("Set-Cookie"), "ut-path")

	// match 4.2
	handler = Interceptor(WithCookieDomain("ut-domain"))
	ctx = newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Contains(t, ctx.Writer.Header().Get("Set-Cookie"), "ut-domain")

	// match 4.3
	handler = Interceptor(WithCookieSameSite(http.SameSiteStrictMode))
	ctx = newCtx(http.MethodGet)
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Contains(t, ctx.Writer.Header().Get("Set-Cookie"), "Strict")
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

func newCtx(method string) *gin.Context {
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		Method: method,
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

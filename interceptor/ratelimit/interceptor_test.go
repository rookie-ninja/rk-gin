package rkginlimit

import (
	"fmt"
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

func TestInterceptor_WithoutOptions(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor()

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}

	handler(ctx)

	assert.False(t, ctx.IsAborted())
}

func TestInterceptor_WithTokenBucket(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithAlgorithm(TokenBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}

	handler(ctx)

	assert.False(t, ctx.IsAborted())
}

func TestInterceptor_WithLeakyBucket(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithAlgorithm(LeakyBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}

	handler(ctx)

	assert.False(t, ctx.IsAborted())
}

func TestInterceptor_WithUserLimiter(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithGlobalLimiter(func(ctx *gin.Context) error {
			return fmt.Errorf("ut-error")
		}),
		WithLimiterByPath("/ut-path", func(ctx *gin.Context) error {
			return fmt.Errorf("ut-error")
		}))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}

	handler(ctx)

	assert.Equal(t, http.StatusTooManyRequests, ctx.Writer.Status())
	assert.False(t, ctx.IsAborted())
}

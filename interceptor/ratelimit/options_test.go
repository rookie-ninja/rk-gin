package rkginlimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestWithEntryNameAndType(t *testing.T) {
	defer assertNotPanic(t)

	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
	assert.Len(t, set.limiter, 1)

	// Should be noop limiter
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	set.getLimiter("")(ctx)
}

func TestWithReqPerSec(t *testing.T) {
	// With non-zero
	set := newOptionSet(
		WithReqPerSec(1))

	assert.Equal(t, 1, set.reqPerSec)
	assert.Len(t, set.limiter, 1)

	// Should be token based limiter
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	set.getLimiter("")(ctx)

	// With zero
	set = newOptionSet(
		WithReqPerSec(0))

	assert.Equal(t, 0, set.reqPerSec)
	assert.Len(t, set.limiter, 1)

	// should be zero rate limiter which returns error
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	assert.NotNil(t, set.getLimiter("")(ctx))
}

func TestWithReqPerSecByPath(t *testing.T) {
	// with non-zero
	set := newOptionSet(
		WithReqPerSecByPath("/ut-path", 1))

	assert.Equal(t, 1, set.reqPerSecByPath["/ut-path"])
	assert.NotNil(t, set.limiter["/ut-path"])

	// Should be token based limiter
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	set.getLimiter("/ut-path")(ctx)

	// With zero
	set = newOptionSet(
		WithReqPerSecByPath("/ut-path", 0))

	assert.Equal(t, 0, set.reqPerSecByPath["/ut-path"])
	assert.NotNil(t, set.limiter["/ut-path"])

	// should be zero rate limiter which returns error
	ctx, _ = gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	assert.NotNil(t, set.getLimiter("/ut-path")(ctx))
}

func TestWithAlgorithm(t *testing.T) {
	defer assertNotPanic(t)

	// With invalid algorithm
	set := newOptionSet(
		WithAlgorithm("invalid-algo"))

	// should be noop limiter
	assert.Len(t, set.limiter, 1)
	set.getLimiter("")

	// With leaky bucket non zero
	set = newOptionSet(
		WithAlgorithm(LeakyBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-method", 1))

	// should be leaky bucket
	assert.Len(t, set.limiter, 2)
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	set.getLimiter("")(ctx)
	set.getLimiter("ut-path")(ctx)
}

func TestWithGlobalLimiter(t *testing.T) {
	set := newOptionSet(
		WithGlobalLimiter(func(ctx *gin.Context) error {
			return fmt.Errorf("ut error")
		}))

	assert.Len(t, set.limiter, 1)
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	assert.NotNil(t, set.getLimiter("")(ctx))
}

func TestWithLimiterByPath(t *testing.T) {
	set := newOptionSet(
		WithLimiterByPath("/ut-path", func(ctx *gin.Context) error {
			return fmt.Errorf("ut error")
		}))

	assert.Len(t, set.limiter, 2)

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	assert.NotNil(t, set.getLimiter("/ut-path")(ctx))
}

func TestOptionSet_Wait(t *testing.T) {
	defer assertNotPanic(t)

	// With user limiter
	set := newOptionSet(
		WithGlobalLimiter(func(*gin.Context) error {
			return nil
		}))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/ut-path",
		},
	}
	set.Wait(ctx, "/ut-path")

	// With token bucket and global limiter
	set = newOptionSet(
		WithAlgorithm(TokenBucket))

	set.Wait(ctx, "/ut-path")

	// With token bucket and limiter by method
	set = newOptionSet(
		WithAlgorithm(TokenBucket),
		WithReqPerSecByPath("/ut-path", 100))

	set.Wait(ctx, "/ut-path")

	// With leaky bucket and global limiter
	set = newOptionSet(
		WithAlgorithm(LeakyBucket))

	set.Wait(ctx, "/ut-path")

	// With leaky bucket and limiter by method
	set = newOptionSet(
		WithAlgorithm(LeakyBucket),
		WithReqPerSecByPath("/ut-path", 100))

	set.Wait(ctx, "/ut-path")

	// Without any configuration
	set = newOptionSet()
	set.Wait(ctx, "/ut-path")
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

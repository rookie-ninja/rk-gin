// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(prometheus.NewRegistry()))

	// With ignoring case
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/assets",
		},
	}
	handler(ctx)

	// Happy case
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Method: "ut-method",
	}
	handler(ctx)
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

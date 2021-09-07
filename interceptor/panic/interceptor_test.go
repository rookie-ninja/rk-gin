// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginpanic

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"os"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"))
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})

	// call interceptor
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

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

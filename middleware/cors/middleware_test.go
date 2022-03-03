// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgincors

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	return ctx
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	beforeCtx := rkmidcors.NewBeforeCtx()
	beforeCtx.Output.HeadersToReturn["key"] = "value"
	beforeCtx.Output.HeaderVary = []string{"vary"}
	mock := rkmidcors.NewOptionSetMock(beforeCtx)

	// case 1: abort
	inter := Middleware(rkmidcors.WithMockOptionSet(mock))
	ctx := newCtx()
	beforeCtx.Output.Abort = true
	inter(ctx)
	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
	assert.Equal(t, "vary", ctx.Writer.Header().Get(rkmid.HeaderVary))

	// case 2: happy case
	ctx = newCtx()
	beforeCtx.Output.Abort = false
	inter(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
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

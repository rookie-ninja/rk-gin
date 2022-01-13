// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgincsrf

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/csrf"
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

	beforeCtx := rkmidcsrf.NewBeforeCtx()
	mock := rkmidcsrf.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Interceptor(rkmidcsrf.WithMockOptionSet(mock))
	ctx := newCtx()

	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.New(rkerror.WithHttpCode(http.StatusForbidden))
	inter(ctx)
	assert.Equal(t, http.StatusForbidden, ctx.Writer.Status())

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	beforeCtx.Output.VaryHeaders = []string{"value"}
	beforeCtx.Output.Cookie = &http.Cookie{}
	ctx = newCtx()
	inter(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.NotEmpty(t, ctx.Writer.Header().Get(rkmid.HeaderVary))
	assert.NotNil(t, ctx.Writer.Header().Get("Set-Cookie"))
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

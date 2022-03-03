// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginauth

import (
	"github.com/gin-gonic/gin"
	rkerror "github.com/rookie-ninja/rk-entry/error"
	"github.com/rookie-ninja/rk-entry/middleware/auth"
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
	beforeCtx := rkmidauth.NewBeforeCtx()
	mock := rkmidauth.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Middleware(rkmidauth.WithMockOptionSet(mock))
	ctx := newCtx()
	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.NewUnauthorized()
	beforeCtx.Output.HeadersToReturn["key"] = "value"
	inter(ctx)
	assert.Equal(t, http.StatusUnauthorized, ctx.Writer.Status())
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	ctx = newCtx()
	inter(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

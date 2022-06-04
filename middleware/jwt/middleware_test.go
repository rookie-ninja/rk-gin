// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginjwt

import (
	"github.com/gin-gonic/gin"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
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
	beforeCtx := rkmidjwt.NewBeforeCtx()
	mock := rkmidjwt.NewOptionSetMock(beforeCtx)
	inter := Middleware(rkmidjwt.WithMockOptionSet(mock))

	// case 1: error response
	beforeCtx.Output.ErrResp = rkmid.GetErrorBuilder().New(http.StatusUnauthorized, "")
	ctx := newCtx()
	inter(ctx)
	assert.Equal(t, http.StatusUnauthorized, ctx.Writer.Status())

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	ctx = newCtx()
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

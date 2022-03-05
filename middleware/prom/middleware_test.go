// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginprom

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)
	beforeCtx := rkmidprom.NewBeforeCtx()
	afterCtx := rkmidprom.NewAfterCtx()
	mock := rkmidprom.NewOptionSetMock(beforeCtx, afterCtx)
	inter := Middleware(rkmidprom.WithMockOptionSet(mock))

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/ut-path", nil)

	inter(ctx)

	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	rkmidprom.ClearAllMetrics()
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

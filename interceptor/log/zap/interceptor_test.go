// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func TestInterceptor_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	beforeCtx := rkmidlog.NewBeforeCtx()
	afterCtx := rkmidlog.NewAfterCtx()
	mock := rkmidlog.NewOptionSetMock(beforeCtx, afterCtx)
	inter := Interceptor(rkmidlog.WithMockOptionSet(mock))
	ctx := newCtx()

	// happy case
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()
	beforeCtx.Output.Event = event
	beforeCtx.Output.Logger = logger

	inter(ctx)

	eventFromCtx, _ := ctx.Get(rkmid.EventKey.String())
	loggerFromCtx, _ := ctx.Get(rkmid.LoggerKey.String())
	assert.Equal(t, event, eventFromCtx.(rkquery.Event))
	assert.Equal(t, logger, loggerFromCtx.(*zap.Logger))

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

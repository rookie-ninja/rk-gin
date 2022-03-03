// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginmeta

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/middleware/meta"
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
	beforeCtx := rkmidmeta.NewBeforeCtx()
	mock := rkmidmeta.NewOptionSetMock(beforeCtx)

	inter := Middleware(rkmidmeta.WithMockOptionSet(mock))
	ctx := newCtx()

	beforeCtx.Input.Event = rkentry.EventEntryNoop.EventFactory.CreateEventNoop()
	beforeCtx.Output.HeadersToReturn["key"] = "value"

	inter(ctx)

	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintrace

import (
	"context"
	"github.com/gin-gonic/gin"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
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

	beforeCtx := rkmidtrace.NewBeforeCtx()
	afterCtx := rkmidtrace.NewAfterCtx()
	mock := rkmidtrace.NewOptionSetMock(beforeCtx, afterCtx, nil, nil, nil)
	beforeCtx.Output.NewCtx = context.TODO()

	// case 1: with error response
	inter := Middleware(rkmidtrace.WithMockOptionSet(mock))
	ctx := newCtx()

	inter(ctx)

	// case 2: happy case
	noopTracerProvider := trace.NewNoopTracerProvider()
	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(ctx, "noop-span")
	beforeCtx.Output.Span = span

	inter(ctx)

	spanFromCtx, exist := ctx.Get(rkmid.SpanKey.String())
	assert.True(t, exist)
	assert.Equal(t, span, spanFromCtx)
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

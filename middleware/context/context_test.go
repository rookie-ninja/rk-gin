// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginctx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestGetIncomingHeaders(t *testing.T) {
	header := http.Header{}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
		Header: header,
	}

	assert.Equal(t, header, GetIncomingHeaders(ctx))
}

func TestGormCtx(t *testing.T) {
	assert.NotNil(t, GormCtx(&gin.Context{}))
}

func TestAddHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context
	AddHeaderToClient(nil, "", "")

	// With nil writer
	ctx := &gin.Context{}
	AddHeaderToClient(ctx, "", "")

	// Happy case
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	AddHeaderToClient(ctx, "key", "value")
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
}

func TestSetHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context
	SetHeaderToClient(nil, "", "")

	// With nil writer
	ctx := &gin.Context{}
	SetHeaderToClient(ctx, "", "")

	// Happy case
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	SetHeaderToClient(ctx, "key", "value")
	assert.Equal(t, "value", ctx.Writer.Header().Get("key"))
}

func TestGetEvent(t *testing.T) {
	// With nil context
	assert.Equal(t, noopEvent, GetEvent(nil))

	// With no event in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, noopEvent, GetEvent(ctx))

	// Happy case
	event := rkquery.NewEventFactory().CreateEventNoop()
	ctx.Set(rkmid.EventKey.String(), event)
	assert.Equal(t, event, GetEvent(ctx))
}

func TestGetLogger(t *testing.T) {
	// With nil context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(nil))

	// With no logger in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))

	// Happy case
	// Add request id and trace id
	ctx.Writer.Header().Set(rkmid.HeaderRequestId, "ut-request-id")
	ctx.Writer.Header().Set(rkmid.HeaderTraceId, "ut-trace-id")
	ctx.Set(rkmid.LoggerKey.String(), rklogger.NoopLogger)

	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))
}

func TestGetRequestId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetRequestId(nil))

	// With no requestId in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Empty(t, GetRequestId(ctx))

	// Happy case
	ctx.Set(rkmid.HeaderRequestId, "ut-request-id")
	assert.Equal(t, "ut-request-id", GetRequestId(ctx))

	// with nil header
	writer := httptest.NewRecorder()
	writer.HeaderMap = nil
	ctx, _ = gin.CreateTestContext(writer)
	assert.Empty(t, GetRequestId(ctx))
}

func TestGetTraceId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetTraceId(nil))

	// With no traceId in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Empty(t, GetTraceId(ctx))

	// Happy case
	ctx.Set(rkmid.HeaderTraceId, "ut-trace-id")
	assert.Equal(t, "ut-trace-id", GetTraceId(ctx))

	// with nil header
	writer := httptest.NewRecorder()
	writer.HeaderMap = nil
	ctx, _ = gin.CreateTestContext(writer)
	assert.Empty(t, GetRequestId(ctx))
}

func TestGetEntryName(t *testing.T) {
	// With nil context
	assert.Empty(t, GetEntryName(nil))

	// With no entry name in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Empty(t, GetEntryName(ctx))

	// Happy case
	ctx.Set(rkmid.EntryNameKey.String(), "ut-entry-name")
	assert.Equal(t, "ut-entry-name", GetEntryName(ctx))
}

func TestGetTraceSpan(t *testing.T) {
	// With no span in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.NotNil(t, GetTraceSpan(ctx))

	// Happy case
	_, span := noopTracerProvider.Tracer("ut-trace").Start(ctx, "noop-span")
	ctx.Set(rkmid.SpanKey.String(), span)
	assert.Equal(t, span, GetTraceSpan(ctx))
}

func TestGetTracer(t *testing.T) {
	// With nil context
	assert.NotNil(t, GetTracer(nil))

	// With no tracer in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.NotNil(t, GetTracer(ctx))

	// Happy case
	tracer := noopTracerProvider.Tracer("ut-trace")
	ctx.Set(rkmid.TracerKey.String(), tracer)
	assert.Equal(t, tracer, GetTracer(ctx))
}

func TestGetTracerProvider(t *testing.T) {
	// With nil context
	assert.NotNil(t, GetTracerProvider(nil))

	// With no tracer provider in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.NotNil(t, GetTracerProvider(ctx))

	// Happy case
	provider := trace.NewNoopTracerProvider()
	ctx.Set(rkmid.TracerProviderKey.String(), provider)
	assert.Equal(t, provider, GetTracerProvider(ctx))
}

func TestGetTracerPropagator(t *testing.T) {
	// With nil context
	assert.Nil(t, GetTracerPropagator(nil))

	// With no tracer propagator in context
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Nil(t, GetTracerPropagator(ctx))

	// Happy case
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkmid.PropagatorKey.String(), prop)
	assert.Equal(t, prop, GetTracerPropagator(ctx))
}

func TestInjectSpanToHttpRequest(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context and request
	InjectSpanToHttpRequest(nil, nil)

	// Happy case
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkmid.PropagatorKey.String(), prop)
	InjectSpanToHttpRequest(ctx, &http.Request{
		Header: http.Header{},
	})
}

func TestNewTraceSpan(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = &http.Request{}
	assert.NotNil(t, NewTraceSpan(ctx, "ut-span"))
}

func TestEndTraceSpan(t *testing.T) {
	defer assertNotPanic(t)

	// With success
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	span := GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, true)

	// With failure
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	span = GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, false)
}

func TestGetJwtToken(t *testing.T) {
	defer assertNotPanic(t)

	// with nil
	assert.Nil(t, GetJwtToken(nil))

	// With failure
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Nil(t, GetJwtToken(ctx))

	// With success
	ctx.Set(rkmid.JwtTokenKey.String(), &jwt.Token{})
	assert.NotNil(t, GetJwtToken(ctx))
}

func TestGetCsrfToken(t *testing.T) {
	defer assertNotPanic(t)

	// with nil
	assert.Empty(t, GetCsrfToken(nil))

	// With failure
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Empty(t, GetCsrfToken(ctx))

	// With success
	ctx.Set(rkmid.CsrfTokenKey.String(), "value")
	assert.Equal(t, "value", GetCsrfToken(ctx))
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
	code := m.Run()
	os.Exit(code)
}

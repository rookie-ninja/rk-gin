// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginctx

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/rookie-ninja/rk-gin/interceptor/extension"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

var (
	noopTracerProvider = trace.NewNoopTracerProvider()
)

// Add Key values to outgoing header
// It should be used only for common usage
func AddToOutgoingHeader(ctx *gin.Context, key string, value string) {
	if ctx == nil || ctx.Writer == nil {
		return
	}
	header := ctx.Writer.Header()
	header.Add(key, value)
}

// Add request id to outgoing header
//
// The request id would be printed on server's query log and client's query log
func SetRequestIdToOutgoingHeader(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	requestId := rkcommon.GenerateRequestId()

	if len(requestId) > 0 {
		ctx.Writer.Header().Set(rkginextension.RequestIdHeaderKeyDefault, requestId)
	}

	return requestId
}

// Add request id to outgoing metadata
//
// The request id would be printed on server's query log and client's query log
func SetRequestIdToOutgoingHeaderWithValue(ctx *gin.Context, requestId string) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	if len(requestId) > 0 {
		ctx.Writer.Header().Set(rkginextension.RequestIdHeaderKeyDefault, requestId)
	}

	return requestId
}

// Extract takes the call-scoped EventData from middleware.
func GetEvent(ctx *gin.Context) rkquery.Event {
	if ctx == nil {
		return rkquery.NewEventFactory().CreateEventNoop()
	}

	event, ok := ctx.Get(rkginbasic.RkEventKey)

	if !ok {
		return rkquery.NewEventFactory().CreateEventNoop()
	}

	return event.(rkquery.Event)
}

// Extract takes the call-scoped zap logger from middleware.
func GetLogger(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return rklogger.NoopLogger
	}

	logger, ok := ctx.Get(rkginbasic.RkLoggerKey)

	if !ok {
		return rklogger.NoopLogger
	}

	requestId := GetRequestId(ctx)
	traceId := GetTraceId(ctx)

	return logger.(*zap.Logger).With(zap.String("requestId", requestId), zap.String("traceId", traceId))
}

// Extract takes the call-scoped trace provider from middleware.
func GetTracerProvider(ctx *gin.Context) trace.TracerProvider {
	if ctx == nil {
		return noopTracerProvider
	}

	provider, ok := ctx.Get(rkginbasic.RkTracerProviderKey)

	if !ok {
		return noopTracerProvider
	}

	return provider.(trace.TracerProvider)
}

// Extract takes the call-scoped propagator from middleware.
func GetTracerPropagator(ctx *gin.Context) propagation.TextMapPropagator {
	if ctx == nil {
		return nil
	}

	propagator, ok := ctx.Get(rkginbasic.RkPropagatorKey)

	if !ok {
		return nil
	}

	return propagator.(propagation.TextMapPropagator)
}

// Extract takes the call-scoped tracer from middleware.
func GetTracer(ctx *gin.Context) trace.Tracer {
	if ctx == nil {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	tracer, ok := ctx.Get(rkginbasic.RkTracerKey)

	if !ok {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	return tracer.(trace.Tracer)
}

func GetTraceId(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}

	traceId, ok := ctx.Get(rkginbasic.RkTraceIdKey)

	if !ok {
		return ""
	}

	return traceId.(string)
}

// Start a new span
func NewSpan(ctx *gin.Context, name string) trace.Span {
	tracer := GetTracer(ctx)
	newCtx, span := tracer.Start(ctx.Request.Context(), name)
	ctx.Request = ctx.Request.WithContext(newCtx)

	GetEvent(ctx).StartTimer(name)

	return span
}

func EndSpan(ctx *gin.Context, span trace.Span, success bool) {
	if success {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}

	readOnlySpan := span.(sdktrace.ReadOnlySpan)

	GetEvent(ctx).EndTimer(readOnlySpan.Name())
	span.End()
}

func InjectTracerIntoHeader(ctx *gin.Context, header *http.Header) {
	propagator := GetTracerPropagator(ctx)

	if propagator == nil {
		return
	}
	propagator.Inject(ctx.Request.Context(), propagation.HeaderCarrier(*header))
}

// Extract request id from context.
// If user enabled extension interceptor, then a random request Id would e assigned and set to context as value.
// If user called SetRequestIdToOutgoingHeader() or SetRequestIdToOutgoingHeaderWithValue(), then a new request id would
// be updated.
func GetRequestId(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	return ctx.Writer.Header().Get(rkginextension.RequestIdHeaderKeyDefault)
}

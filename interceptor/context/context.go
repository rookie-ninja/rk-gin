// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginctx defines utility functions and variables used by Gin middleware
package rkginctx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

const (
	// RequestIdKey is the header key sent to client
	RequestIdKey = "X-Request-Id"
	// TraceIdKey is the header sent to client
	TraceIdKey = "X-Trace-Id"
)

var (
	noopTracerProvider = trace.NewNoopTracerProvider()
	noopEvent          = rkquery.NewEventFactory().CreateEventNoop()
)

// GetIncomingHeaders extract call-scoped incoming headers
func GetIncomingHeaders(ctx *gin.Context) http.Header {
	return ctx.Request.Header
}

// AddHeaderToClient headers that would be sent to client.
// Values would be merged.
func AddHeaderToClient(ctx *gin.Context, key, value string) {
	if ctx == nil || ctx.Writer == nil {
		return
	}

	header := ctx.Writer.Header()
	header.Add(key, value)
}

// SetHeaderToClient headers that would be sent to client.
// Values would be overridden.
func SetHeaderToClient(ctx *gin.Context, key, value string) {
	if ctx == nil || ctx.Writer == nil {
		return
	}
	header := ctx.Writer.Header()
	header.Set(key, value)
}

// GetEvent extract takes the call-scoped EventData from middleware.
func GetEvent(ctx *gin.Context) rkquery.Event {
	if ctx == nil {
		return noopEvent
	}

	if event, ok := ctx.Get(rkgininter.RpcEventKey); ok {
		return event.(rkquery.Event)
	}

	return noopEvent
}

// GetLogger extract takes the call-scoped zap logger from middleware.
func GetLogger(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return rklogger.NoopLogger
	}

	if logger, ok := ctx.Get(rkgininter.RpcLoggerKey); ok {
		requestId := GetRequestId(ctx)
		traceId := GetTraceId(ctx)
		fields := make([]zap.Field, 0)
		if len(requestId) > 0 {
			fields = append(fields, zap.String("requestId", requestId))
		}
		if len(traceId) > 0 {
			fields = append(fields, zap.String("traceId", traceId))
		}

		return logger.(*zap.Logger).With(fields...)
	}

	return rklogger.NoopLogger
}

// GetRequestId extract request id from context.
// If user enabled meta interceptor, then a random request Id would e assigned and set to context as value.
// If user called AddHeaderToClient() with key of RequestIdKey, then a new request id would be updated.
func GetRequestId(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	return ctx.Writer.Header().Get(RequestIdKey)
}

// GetTraceId extract trace id from context.
func GetTraceId(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	return ctx.Writer.Header().Get(TraceIdKey)
}

// GetEntryName extract entry name from context.
func GetEntryName(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}

	if v, ok := ctx.Get(rkgininter.RpcEntryNameKey); ok {
		return v.(string)
	}

	return ""
}

// GetTraceSpan extract the call-scoped span from context.
func GetTraceSpan(ctx *gin.Context) trace.Span {
	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(ctx, "noop-span")

	if ctx == nil {
		return span
	}

	if v, ok := ctx.Get(rkgininter.RpcSpanKey); ok {
		return v.(trace.Span)
	}

	return span
}

// GetTracer extract the call-scoped tracer from context.
func GetTracer(ctx *gin.Context) trace.Tracer {
	if ctx == nil {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	if v, ok := ctx.Get(rkgininter.RpcTracerKey); ok {
		return v.(trace.Tracer)
	}

	return noopTracerProvider.Tracer("rk-trace-noop")
}

// GetTracerProvider extract the call-scoped tracer provider from context.
func GetTracerProvider(ctx *gin.Context) trace.TracerProvider {
	if ctx == nil {
		return noopTracerProvider
	}

	if v, ok := ctx.Get(rkgininter.RpcTracerProviderKey); ok {
		return v.(trace.TracerProvider)
	}

	return noopTracerProvider
}

// GetTracerPropagator extract takes the call-scoped propagator from middleware.
func GetTracerPropagator(ctx *gin.Context) propagation.TextMapPropagator {
	if ctx == nil {
		return nil
	}

	if v, ok := ctx.Get(rkgininter.RpcPropagatorKey); ok {
		return v.(propagation.TextMapPropagator)
	}

	return nil
}

// InjectSpanToHttpRequest inject span to http request
func InjectSpanToHttpRequest(ctx *gin.Context, req *http.Request) {
	if req == nil {
		return
	}

	newCtx := trace.ContextWithRemoteSpanContext(req.Context(), GetTraceSpan(ctx).SpanContext())
	if propagator := GetTracerPropagator(ctx); propagator != nil {
		propagator.Inject(newCtx, propagation.HeaderCarrier(req.Header))
	}
}

// NewTraceSpan start a new span
func NewTraceSpan(ctx *gin.Context, name string) trace.Span {
	tracer := GetTracer(ctx)
	newCtx, span := tracer.Start(ctx.Request.Context(), name)
	ctx.Request = ctx.Request.WithContext(newCtx)

	GetEvent(ctx).StartTimer(name)

	return span
}

// EndTraceSpan end span
func EndTraceSpan(ctx *gin.Context, span trace.Span, success bool) {
	if success {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}

	span.End()
}

// GetJwtToken return jwt.Token if exists
func GetJwtToken(ctx *gin.Context) *jwt.Token {
	if ctx == nil {
		return nil
	}

	if raw, exist := ctx.Get(rkgininter.RpcJwtTokenKey); exist {
		if res, ok := raw.(*jwt.Token); ok {
			return res
		}
	}

	return nil
}

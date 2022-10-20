// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginctx defines utility functions and variables used by Gin middleware
package rkginctx

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v2/cursor"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
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
	if ctx == nil || ctx.Writer == nil || ctx.Writer.Header() == nil {
		return
	}

	header := ctx.Writer.Header()
	header.Add(key, value)
}

// SetHeaderToClient headers that would be sent to client.
// Values would be overridden.
func SetHeaderToClient(ctx *gin.Context, key, value string) {
	if ctx == nil || ctx.Writer == nil || ctx.Writer.Header() == nil {
		return
	}
	header := ctx.Writer.Header()
	header.Set(key, value)
}

// GetCursor create rkcursor.Cursor instance
func GetCursor(ctx *gin.Context) *rkcursor.Cursor {
	return rkcursor.NewCursor(
		rkcursor.WithLogger(GetLogger(ctx)),
		rkcursor.WithEvent(GetEvent(ctx)),
		rkcursor.WithEntryNameAndType(GetEntryName(ctx), "GinEntry"))
}

// GetEvent extract takes the call-scoped EventData from middleware.
func GetEvent(ctx *gin.Context) rkquery.Event {
	if ctx == nil {
		return noopEvent
	}

	if event, ok := ctx.Get(rkmid.EventKey.String()); ok {
		return event.(rkquery.Event)
	}

	return noopEvent
}

// GetLogger extract takes the call-scoped zap logger from middleware.
func GetLogger(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return rklogger.NoopLogger
	}

	if logger, ok := ctx.Get(rkmid.LoggerKey.String()); ok {
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

func GormCtx(ctx *gin.Context) context.Context {
	res := context.Background()
	res = context.WithValue(res, rkmid.LoggerKey.String(), GetLogger(ctx))
	res = context.WithValue(res, rkmid.EventKey.String(), GetEvent(ctx))
	return res
}

// GetRequestId extract request id from context.
// If user enabled meta interceptor, then a random request Id would e assigned and set to context as value.
// If user called AddHeaderToClient() with key of RequestIdKey, then a new request id would be updated.
func GetRequestId(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil || ctx.Writer.Header() == nil {
		return ""
	}

	return ctx.Writer.Header().Get(rkmid.HeaderRequestId)
}

// GetTraceId extract trace id from context.
func GetTraceId(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil || ctx.Writer.Header() == nil {
		return ""
	}

	return ctx.Writer.Header().Get(rkmid.HeaderTraceId)
}

// GetEntryName extract entry name from context.
func GetEntryName(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}

	if v, ok := ctx.Get(rkmid.EntryNameKey.String()); ok {
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

	if v, ok := ctx.Get(rkmid.SpanKey.String()); ok {
		return v.(trace.Span)
	}

	return span
}

// GetTracer extract the call-scoped tracer from context.
func GetTracer(ctx *gin.Context) trace.Tracer {
	if ctx == nil {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	if v, ok := ctx.Get(rkmid.TracerKey.String()); ok {
		return v.(trace.Tracer)
	}

	return noopTracerProvider.Tracer("rk-trace-noop")
}

// GetTracerProvider extract the call-scoped tracer provider from context.
func GetTracerProvider(ctx *gin.Context) trace.TracerProvider {
	if ctx == nil {
		return noopTracerProvider
	}

	if v, ok := ctx.Get(rkmid.TracerProviderKey.String()); ok {
		return v.(trace.TracerProvider)
	}

	return noopTracerProvider
}

// GetTracerPropagator extract takes the call-scoped propagator from middleware.
func GetTracerPropagator(ctx *gin.Context) propagation.TextMapPropagator {
	if ctx == nil {
		return nil
	}

	if v, ok := ctx.Get(rkmid.PropagatorKey.String()); ok {
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

	if raw, exist := ctx.Get(rkmid.JwtTokenKey.String()); exist {
		if res, ok := raw.(*jwt.Token); ok {
			return res
		}
	}

	return nil
}

// GetCsrfToken return csrf token if exists
func GetCsrfToken(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}

	if raw, ok := ctx.Get(rkmid.CsrfTokenKey.String()); ok {
		if res, ok := raw.(string); ok {
			return res
		}
	}

	return ""
}

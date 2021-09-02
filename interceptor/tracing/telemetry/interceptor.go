// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgintrace

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Create a interceptor with opentelemetry.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)
		ctx.Set(rkgininter.RpcTracerKey, set.Tracer)
		ctx.Set(rkgininter.RpcTracerProviderKey, set.Provider)
		ctx.Set(rkgininter.RpcPropagatorKey, set.Propagator)

		span := before(ctx, set)
		defer span.End()

		ctx.Next()

		after(ctx, span)
	}
}

func before(ctx *gin.Context, set *optionSet) oteltrace.Span {
	opts := []oteltrace.SpanOption{
		oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", ctx.Request)...),
		oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(ctx.Request)...),
		oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, ctx.FullPath(), ctx.Request)...),
		oteltrace.WithAttributes(localeToAttributes()...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}

	// 1: extract tracing info from request header
	spanCtx := oteltrace.SpanContextFromContext(
		set.Propagator.Extract(ctx.Request.Context(), propagation.HeaderCarrier(ctx.Request.Header)))

	spanName := ctx.FullPath()
	if len(spanName) < 1 {
		spanName = "rk-span-default"
	}

	// 2: start new span
	newRequestCtx, span := set.Tracer.Start(
		oteltrace.ContextWithRemoteSpanContext(ctx.Request.Context(), spanCtx),
		spanName, opts...)
	// 2.1: pass the span through the request context
	ctx.Request = ctx.Request.WithContext(newRequestCtx)

	// 3: read trace id, tracer, traceProvider, propagator and logger into event data and gin context
	rkginctx.GetEvent(ctx).SetTraceId(span.SpanContext().TraceID().String())
	ctx.Header(rkginctx.TraceIdKey, span.SpanContext().TraceID().String())

	ctx.Set(rkgininter.RpcSpanKey, span)
	return span
}

func after(ctx *gin.Context, span oteltrace.Span) {
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(ctx.Writer.Status())
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(ctx.Writer.Status())
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)
	if len(ctx.Errors) > 0 {
		span.SetAttributes(attribute.String("errors", ctx.Errors.String()))
	}
}

// Convert locale information into attributes.
func localeToAttributes() []attribute.KeyValue {
	res := []attribute.KeyValue{
		attribute.String(rkgininter.Realm.Key, rkgininter.Realm.String),
		attribute.String(rkgininter.Region.Key, rkgininter.Region.String),
		attribute.String(rkgininter.AZ.Key, rkgininter.AZ.String),
		attribute.String(rkgininter.Domain.Key, rkgininter.Domain.String),
	}

	return res
}

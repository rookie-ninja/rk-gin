// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgintrace is aa middleware of gin framework for recording trace info of RPC
package rkgintrace

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-gin/middleware/context"
)

// Middleware create a interceptor with opentelemetry.
func Middleware(opts ...rkmidtrace.Option) gin.HandlerFunc {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())
		ctx.Set(rkmid.TracerKey.String(), set.GetTracer())
		ctx.Set(rkmid.TracerProviderKey.String(), set.GetProvider())
		ctx.Set(rkmid.PropagatorKey.String(), set.GetPropagator())

		beforeCtx := set.BeforeCtx(ctx.Request, false)
		set.Before(beforeCtx)

		// create request with new context
		ctx.Request = ctx.Request.WithContext(beforeCtx.Output.NewCtx)

		// add to context
		if beforeCtx.Output.Span != nil {
			traceId := beforeCtx.Output.Span.SpanContext().TraceID().String()
			rkginctx.GetEvent(ctx).SetTraceId(traceId)
			ctx.Header(rkmid.HeaderTraceId, traceId)
			ctx.Set(rkmid.SpanKey.String(), beforeCtx.Output.Span)
		}

		ctx.Next()

		afterCtx := set.AfterCtx(ctx.Writer.Status(), "")
		set.After(beforeCtx, afterCtx)
	}
}

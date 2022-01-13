// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginlog is a middleware for gin framework for logging RPC.
package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"strconv"
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...rkmidlog.Option) gin.HandlerFunc {
	set := rkmidlog.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		// call before
		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		ctx.Set(rkmid.EventKey.String(), beforeCtx.Output.Event)
		ctx.Set(rkmid.LoggerKey.String(), beforeCtx.Output.Logger)

		// call next
		ctx.Next()

		// call after
		afterCtx := set.AfterCtx(
			rkginctx.GetRequestId(ctx),
			rkginctx.GetTraceId(ctx),
			strconv.Itoa(ctx.Writer.Status()))
		set.After(beforeCtx, afterCtx)
	}
}

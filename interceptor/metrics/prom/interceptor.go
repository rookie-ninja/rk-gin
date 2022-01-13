// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginmetrics is a middleware for gin framework which record prometheus metrics for RPC
package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/metrics"
	"strconv"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...rkmidmetrics.Option) gin.HandlerFunc {
	set := rkmidmetrics.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(set.BeforeCtx(ctx.Request))

		ctx.Next()

		afterCtx := set.AfterCtx(strconv.Itoa(ctx.Writer.Status()))
		set.After(beforeCtx, afterCtx)
	}
}

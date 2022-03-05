// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginlimit is a middleware of gin framework for adding rate limit in RPC response
package rkginlimit

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
)

// Middleware Add rate limit interceptors.
func Middleware(opts ...rkmidlimit.Option) gin.HandlerFunc {
	set := rkmidlimit.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			ctx.AbortWithStatusJSON(beforeCtx.Output.ErrResp.Err.Code, beforeCtx.Output.ErrResp)
			return
		}

		ctx.Next()
	}
}

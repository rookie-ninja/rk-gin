// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginlimit is a middleware of gin framework for adding rate limit in RPC response
package rkginlimit

import (
	"github.com/gin-gonic/gin"
	rkerror "github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"net/http"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		event := rkginctx.GetEvent(ctx)

		if duration, err := set.Wait(ctx, ctx.Request.URL.Path); err != nil {
			event.SetCounter("rateLimitWaitMs", duration.Milliseconds())
			event.AddErr(err)

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, rkerror.New(
				rkerror.WithHttpCode(http.StatusTooManyRequests),
				rkerror.WithDetails(err)))

			return
		}

		ctx.Next()
	}
}

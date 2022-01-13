// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgincors is a CORS middleware for gin framework
package rkgincors

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/cors"
	"net/http"
)

// Interceptor Add CORS interceptors.
func Interceptor(opts ...rkmidcors.Option) gin.HandlerFunc {
	set := rkmidcors.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Writer.Header().Set(k, v)
		}

		for _, v := range beforeCtx.Output.HeaderVary {
			ctx.Writer.Header().Add(rkmid.HeaderVary, v)
		}

		// case 1: with abort
		if beforeCtx.Output.Abort {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		// case 2: call next
		ctx.Next()
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginsec is a middleware of gin framework for adding secure headers in RPC response
package rkginsec

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
)

// Middleware will add secure headers in http response.
func Middleware(opts ...rkmidsec.Option) gin.HandlerFunc {
	set := rkmidsec.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Writer.Header().Set(k, v)
		}

		ctx.Next()
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginmeta is a middleware of gin framework for adding metadata in RPC response
package rkginmeta

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/meta"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
)

// Interceptor will add common headers as extension style in http response.
func Interceptor(opts ...rkmidmeta.Option) gin.HandlerFunc {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(rkginctx.GetEvent(ctx))
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Header(k, v)
		}

		ctx.Next()
	}
}

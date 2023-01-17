// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginpanic is a middleware of gin framework for recovering from panic
package rkginpanic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/rookie-ninja/rk-gin/v2/middleware/context"
	"net/http"
)

// Middleware returns a gin.HandlerFunc (middleware)
func Middleware(opts ...rkmidpanic.Option) gin.HandlerFunc {
	set := rkmidpanic.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		handlerFunc := func(resp rkerror.ErrorInterface) {
			if ctx.Writer.Size() < 1 {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, resp)
			}
		}
		beforeCtx := set.BeforeCtx(rkginctx.GetEvent(ctx), rkginctx.GetLogger(ctx), handlerFunc)
		set.Before(beforeCtx)

		defer beforeCtx.Output.DeferFunc()

		ctx.Next()
	}
}

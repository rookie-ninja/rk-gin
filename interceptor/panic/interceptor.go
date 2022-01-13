// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginpanic is a middleware of gin framework for recovering from panic
package rkginpanic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"net/http"
)

// Interceptor returns a gin.HandlerFunc (middleware)
func Interceptor(opts ...rkmidpanic.Option) gin.HandlerFunc {
	set := rkmidpanic.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		handlerFunc := func(resp *rkerror.ErrorResp) {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, resp)
		}
		beforeCtx := set.BeforeCtx(rkginctx.GetEvent(ctx), rkginctx.GetLogger(ctx), handlerFunc)
		set.Before(beforeCtx)

		defer beforeCtx.Output.DeferFunc()

		ctx.Next()
	}
}

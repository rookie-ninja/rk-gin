// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgincsrf is a middleware of gin framework for adding csrf in RPC response
package rkgincsrf

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"net/http"
)

// Middleware Add csrf interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/csrf.go
func Middleware(opts ...rkmidcsrf.Option) gin.HandlerFunc {
	set := rkmidcsrf.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			ctx.JSON(beforeCtx.Output.ErrResp.Code(), beforeCtx.Output.ErrResp)
			return
		}

		for _, v := range beforeCtx.Output.VaryHeaders {
			ctx.Writer.Header().Add(rkmid.HeaderVary, v)
		}

		if beforeCtx.Output.Cookie != nil {
			http.SetCookie(ctx.Writer, beforeCtx.Output.Cookie)
		}

		// store token in the context
		ctx.Set(rkmid.CsrfTokenKey.String(), beforeCtx.Input.Token)

		ctx.Next()
	}
}

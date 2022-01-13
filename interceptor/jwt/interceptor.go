// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginjwt is a middleware of gin framework for adding jwt in RPC response
package rkginjwt

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/jwt"
)

// Interceptor Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Interceptor(opts ...rkmidjwt.Option) gin.HandlerFunc {
	set := rkmidjwt.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		beforeCtx := set.BeforeCtx(ctx.Request, nil)
		set.Before(beforeCtx)

		// case 1: error response
		if beforeCtx.Output.ErrResp != nil {
			ctx.AbortWithStatusJSON(beforeCtx.Output.ErrResp.Err.Code,
				beforeCtx.Output.ErrResp)
			return
		}

		// insert into context
		ctx.Set(rkmid.JwtTokenKey.String(), beforeCtx.Output.JwtToken)

		// case 2: call next
		ctx.Next()
	}
}

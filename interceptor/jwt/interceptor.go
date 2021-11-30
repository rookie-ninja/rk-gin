// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginjwt is a middleware of gin framework for adding jwt in RPC response
package rkginjwt

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"net/http"
)

// Interceptor Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Next()
			return
		}

		// extract token from extractor
		var auth string
		var err error
		for _, extractor := range set.extractors {
			// Extract token from extractor, if it's not fail break the loop and
			// set auth
			auth, err = extractor(ctx)
			if err == nil {
				break
			}
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return
		}

		// parse token
		token, err := set.ParseTokenFunc(auth, ctx)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return
		}

		// insert into context
		ctx.Set(rkgininter.RpcJwtTokenKey, token)

		ctx.Next()
	}
}

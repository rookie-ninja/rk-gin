// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginauth is auth middleware for gin framework
package rkginauth

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
)

// Middleware validate bellow authorization.
//
// 1: Basic Auth: The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password.
// 2: API key: An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers.
func Middleware(opts ...rkmidauth.Option) gin.HandlerFunc {
	set := rkmidauth.NewOptionSet(opts...)

	return func(ctx *gin.Context) {
		// add entry name into context
		ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(ctx.Request)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			for k, v := range beforeCtx.Output.HeadersToReturn {
				ctx.Writer.Header().Set(k, v)
			}
			ctx.AbortWithStatusJSON(beforeCtx.Output.ErrResp.Code(), beforeCtx.Output.ErrResp)
			return
		}

		// case 2: authorized, call next
		ctx.Next()
	}
}

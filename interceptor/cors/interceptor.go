// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Package rkgincors is a CORS middleware for echo framework
package rkgincors

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"net/http"
	"strconv"
	"strings"
)

// Interceptor Add CORS interceptors.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	allowMethods := strings.Join(set.AllowMethods, ",")
	allowHeaders := strings.Join(set.AllowHeaders, ",")
	exposeHeaders := strings.Join(set.ExposeHeaders, ",")
	maxAge := strconv.Itoa(set.MaxAge)
	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Next()
			return
		}

		originHeader := ctx.Request.Header.Get(headerOrigin)
		preflight := ctx.Request.Method == http.MethodOptions

		// 1: if no origin header was provided, we will return 204 if request is not a OPTION method
		if originHeader == "" {
			// 1.1: if not a preflight request, then pass through
			if !preflight {
				ctx.Next()
				return
			}

			// 1.2: if it is a preflight request, then return with 204
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 2: origin not allowed, we will return 204 if request is not a OPTION method
		if !set.isOriginAllowed(originHeader) {
			// 2.1: if not a preflight request, then pass through
			if !preflight {
				ctx.AbortWithStatus(http.StatusFound)
				return
			}

			// 2.2: if it is a preflight request, then return with 204
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 3: not a OPTION method
		if !preflight {
			ctx.Writer.Header().Set(headerAccessControlAllowOrigin, originHeader)
			// 3.1: add Access-Control-Allow-Credentials
			if set.AllowCredentials {
				ctx.Writer.Header().Set(headerAccessControlAllowCredentials, "true")
			}
			// 3.2: add Access-Control-Expose-Headers
			if exposeHeaders != "" {
				ctx.Writer.Header().Set(headerAccessControlExposeHeaders, exposeHeaders)
			}
			ctx.Next()
			return
		}

		// 4: preflight request, return 204
		// add related headers including:
		//
		// - Vary
		// - Access-Control-Allow-Origin
		// - Access-Control-Allow-Methods
		// - Access-Control-Allow-Credentials
		// - Access-Control-Allow-Headers
		// - Access-Control-Max-Age
		ctx.Writer.Header().Add(headerVary, headerAccessControlRequestMethod)
		ctx.Writer.Header().Add(headerVary, headerAccessControlRequestHeaders)
		ctx.Writer.Header().Set(headerAccessControlAllowOrigin, originHeader)
		ctx.Writer.Header().Set(headerAccessControlAllowMethods, allowMethods)

		// 4.1: Access-Control-Allow-Credentials
		if set.AllowCredentials {
			ctx.Writer.Header().Set(headerAccessControlAllowCredentials, "true")
		}

		// 4.2: Access-Control-Allow-Headers
		if allowHeaders != "" {
			ctx.Writer.Header().Set(headerAccessControlAllowHeaders, allowHeaders)
		} else {
			h := ctx.Request.Header.Get(headerAccessControlRequestHeaders)
			if h != "" {
				ctx.Writer.Header().Set(headerAccessControlAllowHeaders, h)
			}
		}
		if set.MaxAge > 0 {
			// 4.3: Access-Control-Max-Age
			ctx.Writer.Header().Set(headerAccessControlMaxAge, maxAge)
		}

		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
}

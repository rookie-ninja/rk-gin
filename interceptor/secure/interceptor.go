// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginsec is a middleware of gin framework for adding secure headers in RPC response
package rkginsec

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
)

// Interceptor will add secure headers in http response.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		if set.Skipper(ctx) {
			ctx.Next()
			return
		}

		req := ctx.Request
		res := ctx.Writer

		// Add X-XSS-Protection header
		if set.XSSProtection != "" {
			res.Header().Set(headerXXSSProtection, set.XSSProtection)
		}

		// Add X-Content-Type-Options header
		if set.ContentTypeNosniff != "" {
			res.Header().Set(headerXContentTypeOptions, set.ContentTypeNosniff)
		}

		// Add X-Frame-Options header
		if set.XFrameOptions != "" {
			res.Header().Set(headerXFrameOptions, set.XFrameOptions)
		}

		// Add Strict-Transport-Security header
		if (req.TLS != nil || (req.Header.Get(headerXForwardedProto) == "https")) && set.HSTSMaxAge != 0 {
			subdomains := ""
			if !set.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if set.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			res.Header().Set(headerStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", set.HSTSMaxAge, subdomains))
		}

		// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
		if set.ContentSecurityPolicy != "" {
			if set.CSPReportOnly {
				res.Header().Set(headerContentSecurityPolicyReportOnly, set.ContentSecurityPolicy)
			} else {
				res.Header().Set(headerContentSecurityPolicy, set.ContentSecurityPolicy)
			}
		}

		// Add Referrer-Policy header
		if set.ReferrerPolicy != "" {
			res.Header().Set(headerReferrerPolicy, set.ReferrerPolicy)
		}

		ctx.Next()
	}
}

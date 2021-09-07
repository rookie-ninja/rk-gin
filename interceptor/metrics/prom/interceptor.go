// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginmetrics is a middleware for gin framework which record prometheus metrics for RPC
package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"time"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		// start timer
		startTime := time.Now()

		ctx.Next()

		// end timer
		elapsed := time.Now().Sub(startTime)

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if rkgininter.ShouldLog(ctx) {
			if durationMetrics := GetServerDurationMetrics(ctx); durationMetrics != nil {
				durationMetrics.Observe(float64(elapsed.Nanoseconds()))
			}
			if len(ctx.Errors) > 0 {
				if errorMetrics := GetServerErrorMetrics(ctx); errorMetrics != nil {
					errorMetrics.Inc()
				}
			}
			if resCodeMetrics := GetServerResCodeMetrics(ctx); resCodeMetrics != nil {
				resCodeMetrics.Inc()
			}
		}
	}
}

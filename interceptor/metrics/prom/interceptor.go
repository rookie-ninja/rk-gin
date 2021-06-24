// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"time"
)

//var (
//	DefaultLabelKeys = []string{
//		"entryName",
//		"entryType",
//		"realm",
//		"region",
//		"az",
//		"domain",
//		"instance",
//		"appVersion",
//		"appName",
//		"restMethod",
//		"restPath",
//		"type",
//		"resCode",
//	}
//)
//
//const (
//	ElapsedNano = "elapsedNano"
//	Errors      = "errors"
//	ResCode     = "resCode"
//	unknown     = "unknown"
//)

// Create a new prometheus metrics interceptor with options.
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

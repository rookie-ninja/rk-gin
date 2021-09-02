// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkginlog

import (
	"github.com/gin-gonic/gin"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	rkgininter "github.com/rookie-ninja/rk-gin/interceptor"
	rkginctx "github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	ENCODING_CONSOLE int = 0
	ENCODING_JSON    int = 1
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		before(ctx, set)

		ctx.Next()

		after(ctx)
	}
}

func before(ctx *gin.Context, set *optionSet) {
	var event rkquery.Event
	if rkgininter.ShouldLog(ctx) {
		event = set.eventLoggerEntry.GetEventFactory().CreateEvent(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.EntryName),
			rkquery.WithEntryType(set.EntryType))
	} else {
		event = set.eventLoggerEntry.GetEventFactory().CreateEventNoop()
	}

	event.SetStartTime(time.Now())

	remoteIp, remotePort := rkgininter.GetRemoteAddressSet(ctx)
	// handle remote address
	event.SetRemoteAddr(remoteIp + ":" + remotePort)

	payloads := []zap.Field{
		zap.String("apiPath", ctx.Request.URL.Path),
		zap.String("apiMethod", ctx.Request.Method),
		zap.String("apiQuery", ctx.Request.URL.RawQuery),
		zap.String("apiProtocol", ctx.Request.Proto),
		zap.String("userAgent", ctx.Request.UserAgent()),
	}

	// handle payloads
	event.AddPayloads(payloads...)

	// handle operation
	event.SetOperation(ctx.Request.URL.Path)

	ctx.Set(rkgininter.RpcEventKey, event)
	ctx.Set(rkgininter.RpcLoggerKey, set.ZapLogger)
}

func after(ctx *gin.Context) {
	event := rkginctx.GetEvent(ctx)

	// handle errors
	if len(ctx.Errors) > 0 {
		for i := range ctx.Errors {
			event.AddErr(ctx.Errors[i])
		}
	}

	if requestId := rkginctx.GetRequestId(ctx); len(requestId) > 0 {
		event.SetEventId(requestId)
		event.SetRequestId(requestId)
	}

	if traceId := rkginctx.GetTraceId(ctx); len(traceId) > 0 {
		event.SetTraceId(traceId)
	}

	event.SetResCode(strconv.Itoa(ctx.Writer.Status()))
	event.SetEndTime(time.Now())
	event.Finish()
}

// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginctx

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net"
)

var (
	Realm         = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	Region        = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	AZ            = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	Domain        = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	LocalIp       = zap.String("localIp", rkcommon.GetLocalIP())
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())
)

const (
	RequestIdKeyLowerCase = "Requestid"
	RequestIdKeyDash      = "Request-Id"
	RequestIdKeyUnderline = "Request_Id"
	RequestIdKeyDefault   = RequestIdKeyDash
	RkEventKey            = "rkEvent"
	RkLoggerKey           = "rkLogger"
	RkEntryNameKey        = "rkEntry"
	RkEntryNameValue      = "rkEntry"
	RkEntryTypeValue      = "gin"
)

// Add Key values to outgoing header
// It should be used only for common usage
func AddToOutgoingHeader(ctx *gin.Context, key string, values string) {
	if ctx == nil || ctx.Writer == nil {
		return
	}
	header := ctx.Writer.Header()
	header.Add(key, values)
}

// Add request id to outgoing metadata
//
// The request id would be printed on server's query log and client's query log
func AddRequestIdToOutgoingHeader(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	requestId := rkcommon.GenerateRequestId()

	if len(requestId) > 0 {
		AddToOutgoingHeader(ctx, RequestIdKeyDefault, requestId)
	}

	return requestId
}

// Extract takes the call-scoped EventData from gin_zap middleware.
func GetEvent(ctx *gin.Context) rkquery.Event {
	if ctx == nil {
		return rkquery.NewEventFactory().CreateEventNoop()
	}

	event, ok := ctx.Get(RkEventKey)

	if !ok {
		return rkquery.NewEventFactory().CreateEventNoop()
	}

	return event.(rkquery.Event)
}

// Extract takes the call-scoped zap logger from gin_zap middleware.
func GetLogger(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return rklogger.NoopLogger
	}

	logger, ok := ctx.Get(RkLoggerKey)

	if !ok {
		return rklogger.NoopLogger
	}

	return logger.(*zap.Logger)
}

// Extract request ids from outgoing header with bellow keys
//
// keys:
//   request-id
//   request_id
//   requestid
func GetRequestIdsFromOutgoingHeader(ctx *gin.Context) []string {
	res := make([]string, 0)

	if ctx == nil || ctx.Writer == nil {
		return res
	}

	res = append(res, ctx.Writer.Header().Values(RequestIdKeyDash)...)
	res = append(res, ctx.Writer.Header().Values(RequestIdKeyUnderline)...)
	res = append(res, ctx.Writer.Header().Values(RequestIdKeyLowerCase)...)

	return res
}

// Extract request ids from incoming header with bellow keys
//
// keys:
//   request-id
//   request_id
//   requestid
func GetRequestIdsFromIncomingHeader(ctx *gin.Context) []string {
	res := make([]string, 0)

	if ctx == nil || ctx.Request == nil || ctx.Request.Header == nil {
		return res
	}

	res = append(res, ctx.Request.Header.Values(RequestIdKeyDash)...)
	res = append(res, ctx.Request.Header.Values(RequestIdKeyUnderline)...)
	res = append(res, ctx.Request.Header.Values(RequestIdKeyLowerCase)...)

	return res
}

// Get remote endpoint information set including IP, Port, NetworkType
// We will do as best as we can to determine it
// If fails, then just return default ones
func GetRemoteAddressSet(ctx *gin.Context) []zap.Field {
	remoteIP := "0.0.0.0"
	remotePort := "0"

	if ctx == nil || ctx.Request == nil {
		return []zap.Field{
			zap.String("remoteIp", remoteIP),
			zap.String("remotePort", remotePort),
		}
	}

	var err error
	if remoteIP, remotePort, err = net.SplitHostPort(ctx.Request.RemoteAddr); err != nil {
		return []zap.Field{
			zap.String("remoteIp", "0.0.0.0"),
			zap.String("remotePort", "0"),
		}
	}

	forwardedRemoteIP := ctx.GetHeader("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIP) > 0 {
		if forwardedRemoteIP == "::1" {
			forwardedRemoteIP = "localhost"
		}

		remoteIP = forwardedRemoteIP
	}

	if remoteIP == "::1" {
		remoteIP = "localhost"
	}

	return []zap.Field{
		zap.String("remoteIp", remoteIP),
		zap.String("remotePort", remotePort),
	}
}

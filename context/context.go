// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_inter_context

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	rk_logger "github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
)

const (
	RequestIdKeyLowerCase = "requestid"
	RequestIdKeyDash      = "request-id"
	RequestIdKeyUnderline = "request_id"
	RequestIdKeyDefault   = RequestIdKeyDash
)

// Add Key values to outgoing header
// It should be used only for common usage
func AddToOutgoingHeader(ctx *gin.Context, key string, values string) {
	header := ctx.Writer.Header()
	header.Add(key, values)
}

// Add request id to outgoing metadata
//
// The request id would be printed on server's query log and client's query log
func AddRequestIdToOutgoingHeader(ctx *gin.Context) string {
	requestId := GenerateRequestId()

	if len(requestId) > 0 {
		AddToOutgoingHeader(ctx, RequestIdKeyDefault, requestId)
	}

	return requestId
}

// Extract takes the call-scoped EventData from gin_zap middleware.
func GetEvent(ctx *gin.Context) rk_query.Event {
	event, ok := ctx.Get(rk_gin_inter_logging.RKEventKey)
	if !ok {
		return rk_query.NewEventFactory().CreateEventNoop()
	}

	return event.(rk_query.Event)
}

// Extract takes the call-scoped zap logger from gin_zap middleware.
func GetLogger(ctx *gin.Context) *zap.Logger {
	logger, ok := ctx.Get(rk_gin_inter_logging.RKLoggerKey)
	if !ok {
		return rk_logger.NoopLogger
	}

	return logger.(*zap.Logger)
}

func GetRequestIdsFromOutgoingHeader(ctx *gin.Context) []string {
	res := make([]string, 0)

	res = append(res, ctx.Writer.Header().Values(RequestIdKeyDash)...)
	res = append(res, ctx.Writer.Header().Values(RequestIdKeyUnderline)...)
	res = append(res, ctx.Writer.Header().Values(RequestIdKeyLowerCase)...)

	return res
}

func GetRequestIdsFromIncomingHeader(ctx *gin.Context) []string {
	res := make([]string, 0)

	res = append(res, ctx.Request.Header.Values(RequestIdKeyDash)...)
	res = append(res, ctx.Request.Header.Values(RequestIdKeyUnderline)...)
	res = append(res, ctx.Request.Header.Values(RequestIdKeyLowerCase)...)

	return res
}

// Generate request id based on google/uuid
// UUIDs are based on RFC 4122 and DCE 1.1: Authentication and Security Services.
//
// A UUID is a 16 byte (128 bit) array. UUIDs may be used as keys to maps or compared directly.
func GenerateRequestId() string {
	// Do not use uuid.New() since it would panic if any error occurs
	requestId, err := uuid.NewRandom()

	// Currently, we will return empty string if error occurs
	if err != nil {
		return ""
	}

	return requestId.String()
}

// Generate request id based on google/uuid
// UUIDs are based on RFC 4122 and DCE 1.1: Authentication and Security Services.
//
// A UUID is a 16 byte (128 bit) array. UUIDs may be used as keys to maps or compared directly.
func GenerateRequestIdWithPrefix(prefix string) string {
	// Do not use uuid.New() since it would panic if any error occurs
	requestId, err := uuid.NewRandom()

	// Currently, we will return empty string if error occurs
	if err != nil {
		return ""
	}

	return prefix + "-" + requestId.String()
}

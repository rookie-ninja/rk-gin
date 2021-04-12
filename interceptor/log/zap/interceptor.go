// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var optionsMap = make(map[string]*options)

// LoggingZapInterceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func LoggingZapInterceptor(opts ...Option) gin.HandlerFunc {
	defaultOptions := &options{
		entryName:    rkginctx.RKEntryDefaultName,
		eventFactory: rkquery.NewEventFactory(),
		logger:       rklogger.StdoutLogger,
	}

	for i := range opts {
		opts[i](defaultOptions)
	}

	if _, ok := optionsMap[defaultOptions.entryName]; !ok {
		optionsMap[defaultOptions.entryName] = defaultOptions
	}

	return func(ctx *gin.Context) {
		// insert entry name into context
		if len(ctx.GetString(rkginctx.RKEntryNameKey)) < 1 {
			ctx.Set(rkginctx.RKEntryNameKey, defaultOptions.entryName)
		}

		// start timer
		event := defaultOptions.eventFactory.CreateEvent()
		event.SetStartTime(time.Now())
		// insert event data into context
		ctx.Set(rkginctx.RKEventKey, event)

		incomingRequestIds := rkginctx.GetRequestIdsFromIncomingHeader(ctx)
		// insert logger into context
		ctx.Set(rkginctx.RKLoggerKey, defaultOptions.logger.With(zap.Strings("incoming_request_ids", incomingRequestIds)))

		fields := []zap.Field{
			rkginctx.Realm,
			rkginctx.Region,
			rkginctx.AZ,
			rkginctx.Domain,
			rkginctx.AppVersion,
			rkginctx.LocalIP,
			zap.String("entry_name", defaultOptions.entryName),
			zap.String("api_path", ctx.Request.URL.Path),
			zap.String("api_method", ctx.Request.Method),
			zap.String("api_query", ctx.Request.URL.RawQuery),
			zap.String("api_protocol", ctx.Request.Proto),
			zap.String("user_agent", ctx.Request.UserAgent()),
			zap.Strings("incoming_request_ids", incomingRequestIds),
			zap.Time("start_time", event.GetStartTime()),
		}

		remoteAddressSet := rkginctx.GetRemoteAddressSet(ctx)
		fields = append(fields, remoteAddressSet...)
		event.SetRemoteAddr(remoteAddressSet[0].String + ":" + remoteAddressSet[1].String)
		event.SetOperation(ctx.Request.Method + ":" + ctx.Request.URL.Path)

		// handle rest of interceptors
		ctx.Next()

		endTime := time.Now()
		elapsed := endTime.Sub(event.GetStartTime())

		outgoingRequestIds := rkginctx.GetRequestIdsFromOutgoingHeader(ctx)
		ctx.Set(rkginctx.RKLoggerKey, defaultOptions.logger.With(zap.Strings("outgoing_request_ids", outgoingRequestIds)))

		// handle errors
		if len(ctx.Errors) > 0 {
			event.AddErr(ctx.Errors.Last().Err)
		}

		event.SetResCode(strconv.Itoa(ctx.Writer.Status()))
		fields = append(fields,
			zap.Int("res_code", ctx.Writer.Status()),
			zap.Time("end_time", endTime),
			zap.Int64("elapsed_nano", elapsed.Nanoseconds()),
			zap.Strings("outgoing_request_ids", outgoingRequestIds),
		)

		event.AddFields(fields...)
		if len(event.GetEventId()) < 1 {
			ids := append(incomingRequestIds, outgoingRequestIds...)

			if len(ids) > 0 {
				event.SetEventId(strings.Join(ids, ","))
			}
		}

		event.SetEndTime(endTime)
		event.WriteLog()
	}
}

func GetEventFactory(entryName string) *rkquery.EventFactory {
	if _, ok := optionsMap[entryName]; ok {
		return optionsMap[entryName].eventFactory
	}

	return nil
}

func GetEventFactoryFromContext(ctx *gin.Context) *rkquery.EventFactory {
	entryName := ctx.GetString(rkginctx.RKEntryNameKey)

	return GetEventFactory(entryName)
}

func GetLogger(entryName string) *zap.Logger {
	if _, ok := optionsMap[entryName]; ok {
		return optionsMap[entryName].logger
	}

	return nil
}

func GetLoggerFromContext(ctx *gin.Context) *zap.Logger {
	entryName := ctx.GetString(rkginctx.RKEntryNameKey)

	return GetLogger(entryName)
}

func SetLogger(entryName string, logger *zap.Logger) bool {
	if _, ok := optionsMap[entryName]; ok {
		optionsMap[entryName].logger = logger
		return true
	}

	return false
}

func SetLoggerFromContext(ctx *gin.Context, logger *zap.Logger) bool {
	entryName := ctx.GetString(rkginctx.RKEntryNameKey)

	return SetLogger(entryName, logger)
}

// options which is used while initializing logging interceptor
type options struct {
	entryName    string
	eventFactory *rkquery.EventFactory
	logger       *zap.Logger
}

type Option func(*options)

func WithEventFactory(factory *rkquery.EventFactory) Option {
	return func(opt *options) {
		if factory == nil {
			factory = rkquery.NewEventFactory()
		}
		opt.eventFactory = factory
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opt *options) {
		if logger == nil {
			logger = rklogger.NoopLogger
		}
		opt.logger = logger
	}
}

func WithEntryName(entryName string) Option {
	return func(opt *options) {
		opt.entryName = entryName
	}
}

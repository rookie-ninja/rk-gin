// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginlog

import (
	"github.com/gin-gonic/gin"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

// LoggingZapInterceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func LoggingZapInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName:    rkginctx.RkEntryNameValue,
		EntryType:    rkginctx.RkEntryTypeValue,
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
	}

	for i := range opts {
		opts[i](set)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return func(ctx *gin.Context) {
		// start timer
		event := set.EventFactory.CreateEvent(
			rkquery.WithEntryName(set.EntryName),
			rkquery.WithEntryType(set.EntryType))

		event.SetStartTime(time.Now())
		// insert event data into context
		ctx.Set(rkginctx.RkEventKey, event)

		incomingRequestIds := rkginctx.GetRequestIdsFromIncomingHeader(ctx)
		// insert logger into context
		ctx.Set(rkginctx.RkLoggerKey, set.Logger.With(
			zap.String("entryName", set.EntryName),
			zap.String("entryType", set.EntryType),
			zap.Strings("incomingRequestIds", incomingRequestIds)))

		fields := []zap.Field{
			rkginctx.Realm,
			rkginctx.Region,
			rkginctx.AZ,
			rkginctx.Domain,
			zap.String("appVersion", rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			zap.String("appName", rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkginctx.LocalIp,
			zap.String("entryName", set.EntryName),
			zap.String("entryType", set.EntryType),
			zap.String("apiPath", ctx.Request.URL.Path),
			zap.String("apiMethod", ctx.Request.Method),
			zap.String("apiQuery", ctx.Request.URL.RawQuery),
			zap.String("apiProtocol", ctx.Request.Proto),
			zap.String("userAgent", ctx.Request.UserAgent()),
			zap.Strings("incomingRequestIds", incomingRequestIds),
			zap.Time("startTime", event.GetStartTime()),
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
		ctx.Set(rkginctx.RkLoggerKey, set.Logger.With(zap.Strings("outgoingRequestIds", outgoingRequestIds)))

		// handle errors
		errors := ctx.Errors
		if len(errors) > 0 {
			for i := range errors {
				event.AddErr(errors[i])
			}
		}

		event.SetResCode(strconv.Itoa(ctx.Writer.Status()))
		fields = append(fields,
			zap.Int("resCode", ctx.Writer.Status()),
			zap.Time("endTime", endTime),
			zap.Int64("elapsedNano", elapsed.Nanoseconds()),
			zap.Strings("outgoingRequestIds", outgoingRequestIds),
		)

		event.AddFields(fields...)
		if len(event.GetEventId()) < 1 {
			ids := append(incomingRequestIds, outgoingRequestIds...)

			if len(ids) > 0 {
				event.SetEventId(strings.Join(ids, ","))
			}
		}

		event.SetEndTime(endTime)

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if !strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/assets") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/tv") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/sw/") {
			event.WriteLog()
		}
	}
}

func GetEventFactory(ctx *gin.Context) *rkquery.EventFactory {
	if set := getOptionSet(ctx); set != nil {
		return set.EventFactory
	}

	return nil
}

func GetLogger(ctx *gin.Context) *zap.Logger {
	if set := getOptionSet(ctx); set != nil {
		return set.Logger
	}

	return nil
}

func SetLogger(ctx *gin.Context, logger *zap.Logger) bool {
	if set := getOptionSet(ctx); set != nil {
		set.Logger = logger
		return true
	}

	return false
}

func getOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkginctx.RkEntryNameKey)
	return optionsMap[entryName]
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing logging interceptor
type optionSet struct {
	EntryName    string
	EntryType    string
	EventFactory *rkquery.EventFactory
	Logger       *zap.Logger
}

type Option func(*optionSet)

func WithEventFactory(factory *rkquery.EventFactory) Option {
	return func(opt *optionSet) {
		if factory == nil {
			factory = rkquery.NewEventFactory()
		}
		opt.EventFactory = factory
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opt *optionSet) {
		if logger == nil {
			logger = rklogger.NoopLogger
		}
		opt.Logger = logger
	}
}

func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

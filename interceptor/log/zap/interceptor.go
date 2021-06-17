// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/rookie-ninja/rk-gin/interceptor/extension"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
	"time"
)

// LoggingZapInterceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func LoggingZapInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName:    rkginbasic.RkEntryNameValue,
		EntryType:    rkginbasic.RkEntryTypeValue,
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
		BasicFields:  make(map[string]zap.Field),
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
		ctx.Set(rkginbasic.RkEventKey, event)

		// insert logger into context
		ctx.Set(rkginbasic.RkLoggerKey, set.Logger)

		payloads := []zap.Field{
			zap.String("apiPath", ctx.Request.URL.Path),
			zap.String("apiMethod", ctx.Request.Method),
			zap.String("apiQuery", ctx.Request.URL.RawQuery),
			zap.String("apiProtocol", ctx.Request.Proto),
			zap.String("userAgent", ctx.Request.UserAgent()),
		}

		// handle payloads
		event.AddPayloads(payloads...)

		// handle remote address
		remoteAddressSet := getRemoteAddressSet(ctx)
		event.SetRemoteAddr(remoteAddressSet[0].String + ":" + remoteAddressSet[1].String)

		// handle operation
		event.SetOperation(ctx.Request.URL.Path)

		// handle rest of interceptors
		ctx.Next()

		event.SetEndTime(time.Now())

		// handle errors
		if len(ctx.Errors) > 0 {
			for i := range ctx.Errors {
				event.AddErr(ctx.Errors[i])
			}
		}

		// handle resCode
		event.SetResCode(strconv.Itoa(ctx.Writer.Status()))

		// handle requestId and eventId
		requestId := getRequestIdsFromOutgoingHeader(ctx)
		if len(requestId) > 0 {
			event.SetEventId(requestId)
			event.SetRequestId(requestId)
		}

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if !strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/assets") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/rk/v1/tv") &&
			!strings.HasPrefix(ctx.Request.RequestURI, "/sw/") {
			event.Finish()
		}
	}
}

// Extract request id from outgoing header with bellow keys from rkginextension optionset.
// If rkginextension middleware is not enabled, we will use X-RK-RequestId as default key.
func getRequestIdsFromOutgoingHeader(ctx *gin.Context) string {
	if ctx == nil || ctx.Writer == nil {
		return ""
	}

	return ctx.Writer.Header().Get(rkginextension.RequestIdHeaderKeyDefault)
}

// Get remote endpoint information set including IP, Port, NetworkType
// We will do as best as we can to determine it
// If fails, then just return default ones
func getRemoteAddressSet(ctx *gin.Context) []zap.Field {
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

// Get rkquery.EventFactory with this interceptor with gin.Context
func GetEventFactory(ctx *gin.Context) *rkquery.EventFactory {
	if set := GetOptionSet(ctx); set != nil {
		return set.EventFactory
	}

	return nil
}

// Get rkquery.EventFactory with this interceptor with gin.Context
func GetLogger(ctx *gin.Context) *zap.Logger {
	if set := GetOptionSet(ctx); set != nil {
		return set.Logger
	}

	return rklogger.NoopLogger
}

// Get rkquery.Event with this interceptor with gin.Context
func GetEvent(ctx *gin.Context) rkquery.Event {
	if v, ok := ctx.Get(rkginbasic.RkEventKey); !ok {
		return rkquery.NewEventFactory().CreateEventNoop()
	} else {
		return v.(rkquery.Event)
	}
}

// Get rkquery.Event with this interceptor with gin.Context
func SetLogger(ctx *gin.Context, logger *zap.Logger) bool {
	if set := GetOptionSet(ctx); set != nil {
		set.Logger = logger
		return true
	}

	return false
}

// Get optionSet with gin.Context
func GetOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkginbasic.RkEntryNameKey)
	return optionsMap[entryName]
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing logging interceptor
type optionSet struct {
	EntryName    string
	EntryType    string
	EventFactory *rkquery.EventFactory
	Logger       *zap.Logger
	BasicFields  map[string]zap.Field
}

type Option func(*optionSet)

// Provide rkquery.EventFactory.
func WithEventFactory(factory *rkquery.EventFactory) Option {
	return func(opt *optionSet) {
		if factory == nil {
			factory = rkquery.NewEventFactory()
		}
		opt.EventFactory = factory
	}
}

// Provide zap.Logger.
func WithLogger(logger *zap.Logger) Option {
	return func(opt *optionSet) {
		if logger == nil {
			logger = rklogger.NoopLogger
		}
		opt.Logger = logger
	}
}

// Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

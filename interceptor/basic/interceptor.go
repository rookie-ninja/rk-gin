package rkginbasic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
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
	RkEventKey          = "rkEvent"
	RkLoggerKey         = "rkLogger"
	RkTracerKey         = "rkTracer"
	RkTracerProviderKey = "rkTracerProvider"
	RkPropagatorKey     = "rkPropagator"
	RkTraceIdKey        = "rkTraceId"
	RkEntryNameKey      = "rkEntry"
	RkEntryNameValue    = "rkEntry"
	RkEntryTypeValue    = "gin"
)

func BasicInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName: RkEntryNameValue,
		EntryType: RkEntryTypeValue,
	}

	for i := range opts {
		opts[i](set)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return func(ctx *gin.Context) {
		if len(ctx.GetString(RkEntryNameKey)) < 1 {
			ctx.Set(RkEntryNameKey, set.EntryName)
		}

		ctx.Next()
	}
}

func getOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(RkEntryNameKey)
	return optionsMap[entryName]
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing logging interceptor
type optionSet struct {
	EntryName string
	EntryType string
}

type Option func(*optionSet)

func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-prom"
	"strconv"
	"time"
)

var (
	DefaultLabelKeys = []string{
		"realm",
		"region",
		"az",
		"domain",
		"instance",
		"app_version",
		"app_name",
		"method",
		"path",
		"res_code",
	}
)

const (
	ElapsedNano = "elapsed_nano"
	Errors      = "errors"
	ResCode     = "res_code"
	null        = "null"
)

// Global map stores metrics sets
// Interceptor would distinguish metrics set based on
var optionsMap = make(map[string]*options)

func MetricsPromInterceptor(opts ...Option) gin.HandlerFunc {
	defaultOptions := &options{
		entryName:  rkginctx.RKEntryDefaultName,
		registerer: prometheus.DefaultRegisterer,
	}

	for i := range opts {
		opts[i](defaultOptions)
	}

	if len(defaultOptions.entryName) > 0 && defaultOptions.registerer != nil {
		defaultOptions.metricsSet = rkprom.NewMetricsSet(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			defaultOptions.entryName,
			defaultOptions.registerer)
	} else {
		defaultOptions.entryName = rkginctx.RKEntryDefaultName
		defaultOptions.registerer = prometheus.DefaultRegisterer
		defaultOptions.metricsSet = rkprom.NewMetricsSet(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			rkginctx.RKEntryDefaultName,
			defaultOptions.registerer)
	}

	if _, ok := optionsMap[defaultOptions.entryName]; !ok {
		optionsMap[defaultOptions.entryName] = defaultOptions
		// init server and client metrics
		initMetrics(defaultOptions)
	}

	return func(ctx *gin.Context) {
		if len(ctx.GetString(rkginctx.RKEntryNameKey)) < 1 {
			ctx.Set(rkginctx.RKEntryNameKey, defaultOptions.entryName)
		}

		// start timer
		startTime := time.Now()

		ctx.Next()

		// end timer
		elapsed := time.Now().Sub(startTime)

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

func initMetrics(opts *options) {
	opts.metricsSet.RegisterSummary(ElapsedNano, rkprom.SummaryObjectives, DefaultLabelKeys...)
	opts.metricsSet.RegisterCounter(Errors, DefaultLabelKeys...)
	opts.metricsSet.RegisterCounter(ResCode, DefaultLabelKeys...)
}

// options which is used while initializing logging interceptor
type options struct {
	entryName  string
	registerer prometheus.Registerer
	metricsSet *rkprom.MetricsSet
}

type Option func(*options)

func WithEntryName(entryName string) Option {
	return func(opt *options) {
		if len(entryName) > 0 {
			opt.entryName = entryName
		}
	}
}

func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(opt *options) {
		if registerer != nil {
			opt.registerer = registerer
		}
	}
}

// metrics
// Server related
func GetServerDurationMetrics(ctx *gin.Context) prometheus.Observer {
	if metricsSet := GetServerMetricsSetFromContext(ctx); metricsSet != nil {
		return metricsSet.GetSummaryWithValues(ElapsedNano, getValuesFromContext(ctx)...)
	}

	return nil
}

func GetServerErrorMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSetFromContext(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(Errors, getValuesFromContext(ctx)...)
	}

	return nil
}

func GetServerResCodeMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSetFromContext(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(ResCode, getValuesFromContext(ctx)...)
	}

	return nil
}

func GetServerMetricsSet(entryName string) *rkprom.MetricsSet {
	if _, ok := optionsMap[entryName]; ok {
		return optionsMap[entryName].metricsSet
	}

	return nil
}

func GetServerMetricsSetFromContext(ctx *gin.Context) *rkprom.MetricsSet {
	if ctx == nil {
		return nil
	}
	entryName := ctx.GetString(rkginctx.RKEntryNameKey)

	return GetServerMetricsSet(entryName)
}

func ListServerMetricsSets() []*rkprom.MetricsSet {
	res := make([]*rkprom.MetricsSet, 0)
	for _, v := range optionsMap {
		res = append(res, v.metricsSet)
	}

	return res
}

// metrics set already set into context
func getValuesFromContext(ctx *gin.Context) []string {
	method, path, status := null, null, null
	if ctx != nil && ctx.Request != nil {
		method = ctx.Request.Method
		if ctx.Request.URL != nil {
			path = ctx.Request.URL.Path
		}

		if ctx.Writer != nil {
			status = strconv.Itoa(ctx.Writer.Status())
		}
	}

	values := []string{
		rkginctx.Realm.String,
		rkginctx.Region.String,
		rkginctx.AZ.String,
		rkginctx.Domain.String,
		rkginctx.LocalHostname.String,
		rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
		rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
		method,
		path,
		status,
	}

	return values
}

func clearAllMetrics() {
	for _, v := range optionsMap {
		v.metricsSet.UnRegisterSummary(ElapsedNano)
		v.metricsSet.UnRegisterCounter(Errors)
		v.metricsSet.UnRegisterCounter(ResCode)
	}

	optionsMap = make(map[string]*options)
}

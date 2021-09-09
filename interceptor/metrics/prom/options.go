// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-prom"
	"strconv"
	"strings"
)

var (
	// DefaultLabelKeys are default labels for prometheus metrics
	DefaultLabelKeys = []string{
		"entryName",
		"entryType",
		"realm",
		"region",
		"az",
		"domain",
		"instance",
		"appVersion",
		"appName",
		"restMethod",
		"restPath",
		"type",
		"resCode",
	}
)

const (
	ElapsedNano = "elapsedNano" // ElapsedNano records RPC duration
	Errors      = "errors"      // Errors records RPC error
	ResCode     = "resCode"     // ResCode records response code
)

// Register bellow metrics into metrics set.
// 1: Request elapsed time with summary.
// 2: Error count with counter.
// 3: ResCode count with counter.
func initMetrics(opts *optionSet) {
	opts.MetricsSet.RegisterSummary(ElapsedNano, rkprom.SummaryObjectives, DefaultLabelKeys...)
	opts.MetricsSet.RegisterCounter(Errors, DefaultLabelKeys...)
	opts.MetricsSet.RegisterCounter(ResCode, DefaultLabelKeys...)
}

// GetServerDurationMetrics server request elapsed metrics.
func GetServerDurationMetrics(ctx *gin.Context) prometheus.Observer {
	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetSummaryWithValues(ElapsedNano, getValues(ctx)...)
	}

	return nil
}

// GetServerErrorMetrics server error metrics.
func GetServerErrorMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(Errors, getValues(ctx)...)
	}

	return nil
}

// GetServerResCodeMetrics server response code metrics.
func GetServerResCodeMetrics(ctx *gin.Context) prometheus.Counter {
	if ctx == nil {
		return nil
	}

	if metricsSet := GetServerMetricsSet(ctx); metricsSet != nil {
		return metricsSet.GetCounterWithValues(ResCode, getValues(ctx)...)
	}

	return nil
}

// GetServerMetricsSet server metrics set.
func GetServerMetricsSet(ctx *gin.Context) *rkprom.MetricsSet {
	if set := getOptionSet(ctx); set != nil {
		return set.MetricsSet
	}

	return nil
}

// ListServerMetricsSets list all server metrics set associate with GinEntry.
func ListServerMetricsSets() []*rkprom.MetricsSet {
	res := make([]*rkprom.MetricsSet, 0)
	for _, v := range optionsMap {
		res = append(res, v.MetricsSet)
	}

	return res
}

// Metrics set already set into context
func getValues(ctx *gin.Context) []string {
	entryName, entryType, method, path, resCode := "", "", "", "", ""
	if ctx != nil && ctx.Request != nil {
		method = ctx.Request.Method
		if ctx.Request.URL != nil {
			path = ctx.Request.URL.Path
		}

		if ctx.Writer != nil {
			resCode = strconv.Itoa(ctx.Writer.Status())
		}
	}

	if set := getOptionSet(ctx); set != nil {
		entryName = set.EntryName
		entryType = set.EntryType
	}

	values := []string{
		entryName,
		entryType,
		rkgininter.Realm.String,
		rkgininter.Region.String,
		rkgininter.AZ.String,
		rkgininter.Domain.String,
		rkgininter.LocalHostname.String,
		rkentry.GlobalAppCtx.GetAppInfoEntry().Version,
		rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
		method,
		path,
		"gin",
		resCode,
	}

	return values
}

// Internal use only.
func clearAllMetrics() {
	for _, v := range optionsMap {
		v.MetricsSet.UnRegisterSummary(ElapsedNano)
		v.MetricsSet.UnRegisterCounter(Errors)
		v.MetricsSet.UnRegisterCounter(ResCode)
	}

	optionsMap = make(map[string]*optionSet)
}

// Global map stores metrics sets
// Interceptor would distinguish metrics set based on
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:  rkgininter.RpcEntryNameValue,
		EntryType:  rkgininter.RpcEntryTypeValue,
		registerer: prometheus.DefaultRegisterer,
	}

	for i := range opts {
		opts[i](set)
	}

	namespace := strings.ReplaceAll(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, "-", "_")
	subSystem := strings.ReplaceAll(set.EntryName, "-", "_")
	set.MetricsSet = rkprom.NewMetricsSet(
		namespace,
		subSystem,
		set.registerer)

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	initMetrics(set)

	return set
}

// Options which is used while initializing logging interceptor
type optionSet struct {
	EntryName  string
	EntryType  string
	registerer prometheus.Registerer
	MetricsSet *rkprom.MetricsSet
}

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		if len(entryName) > 0 {
			opt.EntryName = entryName
		}

		if len(entryType) > 0 {
			opt.EntryType = entryType
		}
	}
}

// WithRegisterer provide prometheus.Registerer.
func WithRegisterer(registerer prometheus.Registerer) Option {
	return func(opt *optionSet) {
		if registerer != nil {
			opt.registerer = registerer
		}
	}
}

// Get optionSet with gin.Context.
func getOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkgininter.RpcEntryNameKey)
	return optionsMap[entryName]
}

// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_log

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-prom"
	"strconv"
)

var (
	serverMetrics    = initMetrics("server")
	clientMetrics    = initMetrics("client")
	defaultLabelKeys = []string{"realm", "region", "az", "domain", "app_version", "app_name", "method", "path", "protocol", "user_agent", "res_code"}
)

const (
	ElapsedNano         = "elapsed_nano"
	Errors              = "errors"
	BytesTransferredIn  = "bytes_transferred_in"
	BytesTransferredOut = "bytes_transferred_out"
	ResCode             = "res_code"
)

func initMetrics(subSystem string) *rk_prom.MetricsSet {
	metricsSet := rk_prom.NewMetricsSet("rk", subSystem)
	metricsSet.RegisterSummary(ElapsedNano, rk_prom.SummaryObjectives, defaultLabelKeys...)
	metricsSet.RegisterCounter(Errors, defaultLabelKeys...)
	metricsSet.RegisterCounter(BytesTransferredIn, defaultLabelKeys...)
	metricsSet.RegisterCounter(BytesTransferredOut, defaultLabelKeys...)
	metricsSet.RegisterCounter(ResCode, defaultLabelKeys...)

	return metricsSet
}

// Server related
func getServerDurationMetrics(ctx *gin.Context) prometheus.Observer {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetSummaryWithValues(ElapsedNano, values...)
}

func getServerErrorMetrics(ctx *gin.Context) prometheus.Counter {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetCounterWithValues(Errors, values...)
}

func getServerResCodeMetrics(ctx *gin.Context) prometheus.Counter {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetCounterWithValues(ResCode, values...)
}

func GetServerMetricsSet() *rk_prom.MetricsSet {
	return serverMetrics
}

func GetClientMetricsSet() *rk_prom.MetricsSet {
	return clientMetrics
}

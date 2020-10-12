// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_inter_logging

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
	elapsedMS           = "elapsed_ms"
	errors              = "errors"
	bytesTransferredIn  = "bytes_transferred_in"
	bytesTransferredOut = "bytes_transferred_out"
	resCode             = "res_code"
)

func initMetrics(subSystem string) *rk_prom.MetricsSet {
	metricsSet := rk_prom.NewMetricsSet("rk", subSystem)
	metricsSet.RegisterSummary(elapsedMS, rk_prom.SummaryObjectives, defaultLabelKeys...)
	metricsSet.RegisterCounter(errors, defaultLabelKeys...)
	metricsSet.RegisterCounter(bytesTransferredIn, defaultLabelKeys...)
	metricsSet.RegisterCounter(bytesTransferredOut, defaultLabelKeys...)
	metricsSet.RegisterCounter(resCode, defaultLabelKeys...)

	return metricsSet
}

// Server related
func getServerDurationMetrics(ctx *gin.Context) prometheus.Observer {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetSummaryWithValues(elapsedMS, values...)
}

func getServerErrorMetrics(ctx *gin.Context) prometheus.Counter {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetCounterWithValues(errors, values...)
}

func getServerResCodeMetrics(ctx *gin.Context) prometheus.Counter {
	values := []string{realm.String, region.String, az.String, domain.String, appVersion.String, appName, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), strconv.Itoa(ctx.Writer.Status())}
	return serverMetrics.GetCounterWithValues(resCode, values...)
}

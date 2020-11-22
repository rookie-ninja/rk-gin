// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/rookie-ninja/rk-common/context"
	"github.com/rookie-ninja/rk-common/info"
	rk_metrics "github.com/rookie-ninja/rk-common/metrics"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/shirou/gopsutil/v3/cpu"
	"math"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
)

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func shutdownWithError(err error) {
	debug.PrintStack()
	glog.Error(err)
	os.Exit(1)
}

// @Summary Healthy
// @Id 1
// @Tags Healthy
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/healthy [get]
func healthy(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	event.AddPair("healthy", "true")

	ctx.JSON(http.StatusOK, gin.H{
		"healthy": true,
	})
}

// @Summary GC
// @Id 2
// @Tags GC
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/gc [get]
func gc(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	before := rk_info.MemStatsToStruct()
	runtime.GC()
	event.AddFields(rk_info.MemStatsToFields()...)
	after := rk_info.MemStatsToStruct()

	ctx.JSON(http.StatusOK, gin.H{
		"mem_stat_before_gc": before,
		"mem_stat_after_gc":  after,
	})
}

// @Summary Info
// @Id 3
// @Tags Info
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/info [get]
func info(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	event.AddFields(rk_info.BasicInfoToFields()...)
	info := rk_info.BasicInfoToStruct()
	info.Region = setDefaultIfEmpty(info.Region, "N/A")
	info.Realm = setDefaultIfEmpty(info.Realm, "N/A")
	info.AZ = setDefaultIfEmpty(info.AZ, "N/A")
	info.Domain = setDefaultIfEmpty(info.Domain, "N/A")

	ctx.JSON(http.StatusOK, gin.H{
		"info": rk_info.BasicInfoToStruct(),
	})
}

// @Summary Config
// @Id 4
// @Tags Config
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/config [get]
func config(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, gin.H{
		"viper": rk_info.ViperConfigToJSON(),
		"rk":    rk_info.RkConfigToJSON(),
	})
}

// @Summary API
// @Id 5
// @Tags API
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/apis [get]
func apis(ctx *gin.Context) {
	type Api struct {
		Name   string `json:"name"`
		Port   uint64 `json:"port"`
		Path   string `json:"path"`
		Method string `json:"method"`
		SWURL  string `json:"sw_url"`
	}

	res := make([]*Api, 0)

	entries := rk_ctx.GlobalAppCtx.ListEntries()

	for i := range entries {
		raw := entries[i]
		if raw.GetType() == "gin" {
			entry := raw.(*GinEntry)
			routes := entry.GetRouter().Routes()
			for j := range routes {
				info := routes[j]
				api := &Api{
					Name:   entry.GetName(),
					Port:   entry.GetPort(),
					Path:   info.Path,
					Method: info.Method,
					SWURL:  constructSWRURL(entry.GetSWEntry()),
				}
				res = append(res, api)
			}
		}
	}

	ctx.JSON(http.StatusOK, res)
}

// @Summary System Stat
// @Id 6
// @Tags System
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/sys [get]
func sys(ctx *gin.Context) {
	var cpuPercentage, memPercentage float64
	cpuStat, _ := cpu.Percent(0, false)
	memStat := rk_info.MemStatsToStruct()
	for i := range cpuStat {
		cpuPercentage = math.Round(cpuStat[i]*100) / 100
	}

	memPercentage = math.Round(memStat.MemPercentage*100) / 100

	ctx.JSON(http.StatusOK, gin.H{
		"cpu_percentage": cpuPercentage,
		"mem_percentage": memPercentage,
		"mem_usage_mb":   memStat.MemAllocByte / (1024 * 1024),
		"up_time":        rk_info.BasicInfoToStruct().UpTimeStr,
	})
}

// @Summary Request Stat
// @Id 7
// @Tags Request
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/req [get]
func req(ctx *gin.Context) {
	vector := rk_gin_log.GetServerMetricsSet().GetSummaryVec(rk_gin_log.ElapsedNano)
	metrics := rk_metrics.GetRequestMetrics(vector)

	// fill missed metrics
	apis := make([]string, 0)
	entries := rk_ctx.GlobalAppCtx.ListEntries()
	for i := range entries {
		raw := entries[i]
		if raw.GetType() == "gin" {
			entry := raw.(*GinEntry)
			routes := entry.GetRouter().Routes()
			for j := range routes {
				info := routes[j]
				apis = append(apis, info.Path)
			}
		}
	}

	for i := range apis {
		if !containsMetrics(apis[i], metrics) {
			metrics = append(metrics, &rk_metrics.ReqMetricsRK {
				Path: apis[i],
				ResCode: make([]*rk_metrics.ResCodeRK, 0),
			})
		}
	}

	ctx.JSON(http.StatusOK, metrics)
}

// @Summary TV
// @Id 8
// @Tags TV
// @version 1.0
// @produce application/json
// @Success 200 string string
// @Router /v1/rk/tv [get]
func tv(ctx *gin.Context) {
	switch item := ctx.Param("item"); item {
	case "/":
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(dashboardHTML))
	case "/api":
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(apiHTML))
	case "/dashboard":
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(dashboardHTML))
	case "/info":
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(infoHTML))
	default:
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(notFoundHTML))
	}
}

// helper function
func constructSWRURL(sw *swEntry) string {
	if sw == nil {
		return "na"
	}

	return fmt.Sprintf("http://localhost:%d%s", sw.port, sw.path)
}

func setDefaultIfEmpty(src, def string) string {
	if len(src) < 1 {
		return def
	}

	return src
}

func containsMetrics(api string, metrics []*rk_metrics.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].Path == api {
			return true
		}
	}

	return false
}

// @title RK Swagger Example
// @version 1.0
// @description This is a common service with rk-gin.
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.basic BasicAuth

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func dummyFuncForSwagger() {}

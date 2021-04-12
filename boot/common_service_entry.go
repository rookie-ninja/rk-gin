// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/shirou/gopsutil/v3/cpu"
	"go.uber.org/zap"
	"math"
	"net/http"
	"runtime"
	"strings"
)

// @title RK Swagger Example
// @version 1.0
// @description This is a common service with rk-gin.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.basic BasicAuth

// @name Authorization

// Response for path of /healthy
type HealthyResponse struct {
	Healthy bool `json:"healthy"`
}

// Response for path of /gc
type GCResponse struct {
	MemStatBeforeGC *rkentry.MemStatsInfo `json:"mem_stat_before_gc"`
	MemStatAfterGC  *rkentry.MemStatsInfo `json:"mem_stat_after_gc"`
}

// Response for path of /config
type ConfigResponse struct {
	ViperConfig string `json:"viper"`
}

// Response for path of /apis
type APIsResponse struct {
	Name   string `json:"name"`
	Port   uint64 `json:"port"`
	Path   string `json:"path"`
	Method string `json:"method"`
	SWURL  string `json:"sw_url"`
}

// Response for path of sys
type SysResponse struct {
	CPUUsagePercentage float64 `json:"cpu_usage_percentage"`
	MemUsagePercentage float64 `json:"mem_usage_percentage"`
	MemUsageMB         uint64  `json:"mem_usage_mb"`
	SysUpTime          string  `json:"sys_up_time"`
}

// Bootstrap config of common service.
// 1: Enabled: Enable common service.
// 2: PathPrefix: Prefix of common service path, default value is /v1/rk/.
type BootConfigCommonService struct {
	Enabled    bool   `yaml:"enabled"`
	PathPrefix string `yaml:"pathPrefix"`
}

// RK common service which contains utility API
// 1: Healthy GET Returns true if process is alive
// 2: GC GET Trigger gc()
// 3: Info GET Returns entry basic information
// 4: Config GET Returns viper configs in GlobalAppCtx
// 5: APIs GET Returns list of apis registered in gin router
// 6: Sys GET Returns CPU and Memory information
// 7: Req GET Returns request metrics
type CommonServiceEntry struct {
	entryName        string
	entryType        string
	EventLoggerEntry *rkentry.EventLoggerEntry
	ZapLoggerEntry   *rkentry.ZapLoggerEntry
	PathPrefix       string
}

// Common service entry option function.
type CommonServiceEntryOption func(*CommonServiceEntry)

// Provide name.
func WithNameCommonService(name string) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.entryName = name
	}
}

// Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryCommonService(eventLoggerEntry *rkentry.EventLoggerEntry) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryCommonService(zapLoggerEntry *rkentry.ZapLoggerEntry) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// Provide prefix of common service api path.
func WithPathPrefixCommonService(pathPrefix string) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		if len(pathPrefix) > 0 {
			entry.PathPrefix = pathPrefix
		}
	}
}

// Create new common service entry with options.
func NewCommonServiceEntry(opts ...CommonServiceEntryOption) *CommonServiceEntry {
	entry := &CommonServiceEntry{
		entryName:        "gin-common-service",
		entryType:        "gin-common-service",
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		PathPrefix:       "/v1/rk/",
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.PathPrefix, "/") {
		entry.PathPrefix = "/" + entry.PathPrefix
	}

	if !strings.HasSuffix(entry.PathPrefix, "/") {
		entry.PathPrefix = entry.PathPrefix + "/"
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "gin-common-service"
	}

	return entry
}

// Bootstrap common service entry
func (entry *CommonServiceEntry) Bootstrap(context.Context) {
	// No op
}

// Interrupt common service entry
func (entry *CommonServiceEntry) Interrupt(context.Context) {}

// Get name of entry
func (entry *CommonServiceEntry) GetName() string {
	return entry.entryName
}

// Get entry type
func (entry *CommonServiceEntry) GetType() string {
	return entry.entryType
}

// Stringfy common service entry
func (entry *CommonServiceEntry) String() string {
	m := map[string]interface{}{
		"entry_name":  entry.entryName,
		"entry_type":  entry.entryType,
		"path_prefix": entry.PathPrefix,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Warn("failed to marshal common service entry to string", zap.Error(err))
		return "{}"
	}

	return string(bytes)
}

// @Summary Check healthy status
// @Id 1
// @version 1.0
// @produce application/json
// @Success 200 {object} HealthyResponse
// @Router /v1/rk/healthy [get]
func (entry *CommonServiceEntry) Healthy(ctx *gin.Context) {
	if ctx == nil {
		return
	}
	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doHealthy(ctx))
}

// @Summary Trigger GC
// @Id 2
// @version 1.0
// @produce application/json
// @Success 200 {object} GCResponse
// @Router /v1/rk/gc [get]
func (entry *CommonServiceEntry) GC(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doGC(ctx))
}

// @Summary Service Info
// @Id 3
// @version 1.0
// @produce application/json
// @Success 200 {object} rkinfo.BasicInfo
// @Router /v1/rk/info [get]
func (entry *CommonServiceEntry) Info(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doInfo(ctx))
}

// @Summary List Configs
// @Id 4
// @version 1.0
// @produce application/json
// @Success 200 {object} ConfigResponse
// @Router /v1/rk/config [get]
func (entry *CommonServiceEntry) Config(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doConfig(ctx))
}

// @Summary List APIs
// @Id 5
// @version 1.0
// @produce application/json
// @Success 200 {array} APIsResponse
// @Router /v1/rk/apis [get]
func (entry *CommonServiceEntry) APIs(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	ctx.Header("Access-Control-Allow-Origin", "*")

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doAPIs(ctx))
}

// @Summary System Stat
// @Id 6
// @version 1.0
// @produce application/json
// @Success 200 {object} SysResponse
// @Router /v1/rk/sys [get]
func (entry *CommonServiceEntry) Sys(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doSys(ctx))
}

// @Summary Request Stat
// @Id 7
// @version 1.0
// @produce application/json
// @success 200 {object} rkmetrics.ReqMetricsRK
// @Router /v1/rk/req [get]
func (entry *CommonServiceEntry) Req(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doReq(ctx))
}

// Helper function for Healthy call
func doHealthy(*gin.Context) *HealthyResponse {
	return &HealthyResponse{
		Healthy: true,
	}
}

// Helper function for GC call
func doGC(*gin.Context) *GCResponse {
	before := rkentry.NewMemStatsInfo()
	runtime.GC()
	after := rkentry.NewMemStatsInfo()

	return &GCResponse{
		MemStatBeforeGC: before,
		MemStatAfterGC:  after,
	}
}

// Helper function for Info call
func doInfo(*gin.Context) *rkentry.ProcessInfo {
	return rkentry.NewProcessInfo()
}

// Helper function for Config call
func doConfig(*gin.Context) *ConfigResponse {
	return &ConfigResponse{
		ViperConfig: rkcommon.ConvertStructToJSON(rkentry.NewViperConfigInfo()),
	}
}

// Helper function for APIs call
func doAPIs(ctx *gin.Context) []*APIsResponse {
	res := make([]*APIsResponse, 0)

	if ctx == nil {
		return res
	}

	ginEntry := getEntry(ctx)

	if ginEntry != nil {
		routes := ginEntry.Router.Routes()
		for j := range routes {
			info := routes[j]
			api := &APIsResponse{
				Name:   ginEntry.GetName(),
				Port:   ginEntry.Port,
				Path:   info.Path,
				Method: info.Method,
				SWURL:  constructSWRURL(ginEntry, ctx),
			}
			res = append(res, api)
		}
	}

	return res
}

// Helper function for Sys call
func doSys(*gin.Context) *SysResponse {
	var cpuUsagePercentage, memUsagePercentage float64
	cpuStat, _ := cpu.Percent(0, false)
	memStat := rkentry.NewMemStatsInfo()
	for i := range cpuStat {
		cpuUsagePercentage = math.Round(cpuStat[i]*100) / 100
	}

	memUsagePercentage = math.Round(memStat.MemPercentage*100) / 100

	return &SysResponse{
		CPUUsagePercentage: cpuUsagePercentage,
		MemUsagePercentage: memUsagePercentage,
		MemUsageMB:         memStat.MemAllocByte / (1024 * 1024),
		SysUpTime:          rkentry.NewProcessInfo().UpTimeStr,
	}
}

// Helper function for Req call
func doReq(ctx *gin.Context) []*rkentry.ReqMetricsRK {
	vector := rkginmetrics.GetServerMetricsSetFromContext(ctx).GetSummary(rkginmetrics.ElapsedNano)
	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// fill missed metrics
	apis := make([]string, 0)

	ginEntry := GetGinEntry(ctx.GetString(rkginctx.RKEntryNameKey))
	if ginEntry != nil {
		routes := ginEntry.Router.Routes()
		for j := range routes {
			info := routes[j]
			apis = append(apis, info.Path)
		}
	}

	// add empty metrics into result
	for i := range apis {
		if !containsMetrics(apis[i], reqMetrics) {
			reqMetrics = append(reqMetrics, &rkentry.ReqMetricsRK{
				Path:    apis[i],
				ResCode: make([]*rkentry.ResCodeRK, 0),
			})
		}
	}

	return reqMetrics
}

// Helper functions
func constructSWRURL(entry *GinEntry, ctx *gin.Context) string {
	if entry == nil || entry.SWEntry == nil {
		return "N/A"
	}

	originalURL := fmt.Sprintf("localhost:%d", entry.Port)
	if ctx != nil && ctx.Request != nil && len(ctx.Request.Host) > 0 {
		originalURL = ctx.Request.Host
	}

	return fmt.Sprintf("http://%s%s", originalURL, entry.SWEntry.Path)
}

func containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].Path == api {
			return true
		}
	}

	return false
}

// Extract Gin entry from gin_zap middleware
func getEntry(ctx *gin.Context) *GinEntry {
	if ctx == nil {
		return nil
	}

	entryRaw := rkentry.GlobalAppCtx.GetEntry(ctx.GetString(rkginctx.RKEntryNameKey))
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GinEntry)
	return entry
}

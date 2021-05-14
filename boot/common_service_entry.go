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
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

const (
	CommonServiceEntryType        = "GinCommonServiceEntry"
	CommonServiceEntryNameDefault = "GinCommonServiceDefault"
	CommonServiceEntryDescription = "Internal RK entry which implements commonly used API with Gin framework."
)

// @title RK Swagger for Gin
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

// Bootstrap config of common service.
// 1: Enabled: Enable common service.
type BootConfigCommonService struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// RK common service which contains commonly used APIs
// 1: Healthy GET Returns true if process is alive
// 2: Gc GET Trigger gc()
// 3: Info GET Returns entry basic information
// 4: Configs GET Returns viper configs in GlobalAppCtx
// 5: Apis GET Returns list of apis registered in gin router
// 6: Sys GET Returns CPU and Memory information
// 7: Req GET Returns request metrics
// 8: Certs GET Returns certificates
// 9: Entries GET Returns entries
type CommonServiceEntry struct {
	EntryName        string                    `json:"entryName" yaml:"entryName"`
	EntryType        string                    `json:"entryType" yaml:"entryType"`
	EntryDescription string                    `json:"entryDescription" yaml:"entryDescription"`
	EventLoggerEntry *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	ZapLoggerEntry   *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
}

// Common service entry option.
type CommonServiceEntryOption func(*CommonServiceEntry)

// Provide name.
func WithNameCommonService(name string) CommonServiceEntryOption {
	return func(entry *CommonServiceEntry) {
		entry.EntryName = name
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

// Create new common service entry with options.
func NewCommonServiceEntry(opts ...CommonServiceEntryOption) *CommonServiceEntry {
	entry := &CommonServiceEntry{
		EntryName:        CommonServiceEntryNameDefault,
		EntryType:        CommonServiceEntryType,
		EntryDescription: CommonServiceEntryDescription,
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
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

	if len(entry.EntryName) < 1 {
		entry.EntryName = CommonServiceEntryNameDefault
	}

	return entry
}

// Bootstrap common service entry.
func (entry *CommonServiceEntry) Bootstrap(context.Context) {
	// No op
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	entry.ZapLoggerEntry.GetLogger().Info("Bootstrapping CommonServiceEntry.", event.GetFields()...)
}

// Interrupt common service entry.
func (entry *CommonServiceEntry) Interrupt(context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	entry.ZapLoggerEntry.GetLogger().Info("Interrupting CommonServiceEntry.", event.GetFields()...)
}

// Get name of entry.
func (entry *CommonServiceEntry) GetName() string {
	return entry.EntryName
}

// Get entry type.
func (entry *CommonServiceEntry) GetType() string {
	return entry.EntryType
}

// Stringfy entry.
func (entry *CommonServiceEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// Get description of entry.
func (entry *CommonServiceEntry) GetDescription() string {
	return entry.EntryDescription
}

// Marshal entry.
func (entry *CommonServiceEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
	}

	return json.Marshal(&m)
}

// Not supported.
func (entry *CommonServiceEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Add basic fields into event.
func (entry *CommonServiceEntry) logBasicInfo(event rkquery.Event) {
	event.AddFields(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
	)
}

// Response of /healthy
type HealthyResponse struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
}

// Helper function of /healthy call
func doHealthy(*gin.Context) *HealthyResponse {
	return &HealthyResponse{
		Healthy: true,
	}
}

// @Summary Get application healthy status
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

// Response of /gc
// Returns memory stats of GC before and after.
type GcResponse struct {
	MemStatBeforeGc *rkentry.MemInfo `json:"memStatBeforeGc" yaml:"memStatBeforeGc"`
	MemStatAfterGc  *rkentry.MemInfo `json:"memStatAfterGc" yaml:"memStatAfterGc"`
}

// Helper function of /gc
func doGc(*gin.Context) *GcResponse {
	before := rkentry.NewMemInfo()
	runtime.GC()
	after := rkentry.NewMemInfo()

	return &GcResponse{
		MemStatBeforeGc: before,
		MemStatAfterGc:  after,
	}
}

// @Summary Trigger Gc
// @Id 2
// @version 1.0
// @produce application/json
// @Success 200 {object} GcResponse
// @Router /v1/rk/gc [get]
func (entry *CommonServiceEntry) Gc(ctx *gin.Context) {
	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	if ctx == nil {
		return
	}

	ctx.JSON(http.StatusOK, doGc(ctx))
}

// Helper function of /info
func doInfo(*gin.Context) *rkentry.ProcessInfo {
	return rkentry.NewProcessInfo()
}

// @Summary Get application and process info
// @Id 3
// @version 1.0
// @produce application/json
// @Success 200 {object} rkentry.ProcessInfo
// @Router /v1/rk/info [get]
func (entry *CommonServiceEntry) Info(ctx *gin.Context) {
	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	if ctx == nil {
		return
	}

	ctx.JSON(http.StatusOK, doInfo(ctx))
}

// Response of /configs
type ConfigResponse struct {
	EntryName        string                 `json:"entryName" yaml:"entryName"`
	EntryType        string                 `json:"entryType" yaml:"entryType"`
	EntryDescription string                 `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        map[string]interface{} `json:"entryMeta" yaml:"entryMeta"`
	Path             string                 `json:"path" yaml:"path"`
}

// Helper function of /configs
func doConfigs(*gin.Context) []*ConfigResponse {
	res := make([]*ConfigResponse, 0)

	for _, v := range rkentry.GlobalAppCtx.ListConfigEntries() {
		res = append(res, &ConfigResponse{
			EntryName:        v.GetName(),
			EntryType:        v.GetType(),
			EntryDescription: v.GetDescription(),
			EntryMeta:        v.GetViperAsMap(),
			Path:             v.Path,
		})
	}

	return res
}

// @Summary List ConfigEntry
// @Id 4
// @version 1.0
// @produce application/json
// @Success 200 {object} ConfigResponse
// @Router /v1/rk/configs [get]
func (entry *CommonServiceEntry) Configs(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doConfigs(ctx))
}

// Response for path of /apis
type ApiResponse struct {
	Name   string `json:"name" yaml:"name"`
	Port   uint64 `json:"port" yaml:"port"`
	Path   string `json:"path" yaml:"path"`
	Method string `json:"method" yaml:"method"`
	SwUrl  string `json:"swUrl" yaml:"swUrl"`
}

// Construct swagger URL based on IP and scheme
func constructSwUrl(entry *GinEntry, ctx *gin.Context) string {
	if entry == nil || entry.SwEntry == nil {
		return "N/A"
	}

	originalURL := fmt.Sprintf("localhost:%d", entry.Port)
	if ctx != nil && ctx.Request != nil && len(ctx.Request.Host) > 0 {
		originalURL = ctx.Request.Host
	}

	scheme := "http"
	if ctx != nil && ctx.Request != nil && ctx.Request.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, originalURL, entry.SwEntry.Path)
}

// Is metrics from prometheus contains particular api?
func containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].Path == api {
			return true
		}
	}

	return false
}

// Helper function for APIs call
func doApis(ctx *gin.Context) []*ApiResponse {
	res := make([]*ApiResponse, 0)

	if ctx == nil {
		return res
	}

	ginEntry := getEntry(ctx)

	if ginEntry != nil {
		routes := ginEntry.Router.Routes()
		for j := range routes {
			info := routes[j]

			api := &ApiResponse{
				Name:   ginEntry.GetName(),
				Port:   ginEntry.Port,
				Path:   info.Path,
				Method: info.Method,
				SwUrl:  constructSwUrl(ginEntry, ctx),
			}
			res = append(res, api)
		}
	}

	return res
}

// @Summary List API
// @Id 5
// @version 1.0
// @produce application/json
// @Success 200 {array} ApiResponse
// @Router /v1/rk/apis [get]
func (entry *CommonServiceEntry) Apis(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	ctx.Header("Access-Control-Allow-Origin", "*")

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doApis(ctx))
}

// Response of /sys
type SysResponse struct {
	CpuInfo   *rkentry.CpuInfo   `json:"cpuInfo" yaml:"cpuInfo"`
	MemInfo   *rkentry.MemInfo   `json:"memInfo" yaml:"memInfo"`
	NetInfo   *rkentry.NetInfo   `json:"netInfo" yaml:"netInfo"`
	OsInfo    *rkentry.OsInfo    `json:"osInfo" yaml:"osInfo"`
	GoEnvInfo *rkentry.GoEnvInfo `json:"goEnvInfo" yaml:"goEnvInfo"`
}

// Helper function of /sys
func doSys(*gin.Context) *SysResponse {
	return &SysResponse{
		CpuInfo:   rkentry.NewCpuInfo(),
		MemInfo:   rkentry.NewMemInfo(),
		NetInfo:   rkentry.NewNetInfo(),
		OsInfo:    rkentry.NewOsInfo(),
		GoEnvInfo: rkentry.NewGoEnvInfo(),
	}
}

// @Summary Get OS Stat
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

// Helper function for Req call
func doReq(ctx *gin.Context) []*rkentry.ReqMetricsRK {
	vector := rkginmetrics.GetServerMetricsSet(ctx).GetSummary(rkginmetrics.ElapsedNano)
	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// Fill missed metrics
	apis := make([]string, 0)

	ginEntry := GetGinEntry(ctx.GetString(rkginctx.RKEntryNameKey))
	if ginEntry != nil {
		routes := ginEntry.Router.Routes()
		for j := range routes {
			info := routes[j]
			apis = append(apis, info.Path)
		}
	}

	// Add empty metrics into result
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

// @Summary List prometheus metrics of requests
// @Id 7
// @version 1.0
// @produce application/json
// @success 200 {object} rkentry.ReqMetricsRK
// @Router /v1/rk/req [get]
func (entry *CommonServiceEntry) Req(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doReq(ctx))
}

// Response of /entries
type EntryResponse struct {
	Entries map[string][]*EntryElement `json:"entries" yaml:"entries"`
}

// Entry element which specifies name, type and description.
type EntryElement struct {
	EntryName        string        `json:"entryName" yaml:"entryName"`
	EntryType        string        `json:"entryType" yaml:"entryType"`
	EntryDescription string        `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        rkentry.Entry `json:"entryMeta" yaml:"entryMeta"`
}

// Helper function of /entries
func doEntriesHelper(m map[string]rkentry.Entry, res *EntryResponse) {
	entries := make([]*EntryElement, 0)

	// Iterate entries and construct EntryElement
	for i := range m {
		entry := m[i]
		element := &EntryElement{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
			EntryMeta:        entry,
		}

		entries = append(entries, element)
	}

	var entryType string

	if len(entries) > 0 {
		entryType = entries[0].EntryType
		res.Entries[entryType] = entries
	}
}

// Helper function of /entries
func doEntries(ctx *gin.Context) *EntryResponse {
	res := &EntryResponse{
		Entries: make(map[string][]*EntryElement),
	}

	if ctx == nil {
		return res
	}
	// Add auto generated request Id
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	// Iterate all internal and external entries in GlobalAppCtx
	doEntriesHelper(rkentry.GlobalAppCtx.ListEntries(), res)
	doEntriesHelper(rkentry.GlobalAppCtx.ListEventLoggerEntriesRaw(), res)
	doEntriesHelper(rkentry.GlobalAppCtx.ListZapLoggerEntriesRaw(), res)
	doEntriesHelper(rkentry.GlobalAppCtx.ListConfigEntriesRaw(), res)
	doEntriesHelper(rkentry.GlobalAppCtx.ListCertEntriesRaw(), res)
	doEntriesHelper(rkentry.GlobalAppCtx.ListEntries(), res)

	// App info entry
	appInfoEntry := rkentry.GlobalAppCtx.GetAppInfoEntry()
	res.Entries[appInfoEntry.GetType()] = []*EntryElement{
		{
			EntryName:        appInfoEntry.GetName(),
			EntryType:        appInfoEntry.GetType(),
			EntryDescription: appInfoEntry.GetDescription(),
			EntryMeta:        appInfoEntry,
		},
	}

	return res
}

// @Summary List all Entry
// @Id 8
// @version 1.0
// @produce application/json
// @Success 200 {array} EntryResponse
// @Router /v1/rk/entries [get]
func (entry *CommonServiceEntry) Entries(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doEntries(ctx))
}

// Response of /certs
type CertResponse struct {
	EntryName        string `json:"entryName" yaml:"entryName"`
	EntryType        string `json:"entryType" yaml:"entryType"`
	EntryDescription string `json:"entryDescription" yaml:"entryDescription"`
	ServerCertPath   string `json:"serverCertPath" yaml:"serverCertPath"`
	ServerKeyPath    string `json:"serverKeyPath" yaml:"serverKeyPath"`
	ClientCertPath   string `json:"clientCertPath" yaml:"clientCertPath"`
	ClientKeyPath    string `json:"clientKeyPath" yaml:"clientKeyPath"`
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	Locale           string `json:"locale" yaml:"locale"`
	Provider         string `json:"provider" yaml:"provider"`
	ServerCert       string `json:"serverCert" yaml:"serverCert"`
	ClientCert       string `json:"clientCert" yaml:"clientCert"`
}

// Helper function of /entries
func doCerts(ctx *gin.Context) []*CertResponse {
	res := make([]*CertResponse, 0)

	if ctx == nil {
		return res
	}
	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	entries := rkentry.GlobalAppCtx.ListCertEntries()

	// Iterator cert entries and construct CertResponse
	for i := range entries {
		entry := entries[i]

		response := &CertResponse{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
		}

		if entry.Retriever != nil {
			response.Endpoint = entry.Retriever.GetEndpoint()
			response.Locale = entry.Retriever.GetLocale()
			response.Provider = entry.Retriever.GetProvider()
			response.ServerCertPath = entry.Retriever.GetServerCertPath()
			response.ServerKeyPath = entry.Retriever.GetServerKeyPath()
			response.ClientCertPath = entry.Retriever.GetClientCertPath()
			response.ClientKeyPath = entry.Retriever.GetClientKeyPath()
		}

		if entry.Store != nil {
			response.ServerCert = entry.Store.SeverCertString()
			response.ClientCert = entry.Store.ClientCertString()
		}

		res = append(res, response)
	}

	return res
}

// @Summary List CertEntry
// @Id 9
// @version 1.0
// @produce application/json
// @Success 200 {array} CertResponse
// @Router /v1/rk/certs [get]
func (entry *CommonServiceEntry) Certs(ctx *gin.Context) {
	if ctx == nil {
		return
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doCerts(ctx))
}

// Response of /logs.
type LogResponse struct {
	Entries map[string][]*LogEntryElement `json:"entries" yaml:"entries"`
}

// Entry element which specifies name, type. description, output path and error output path.
type LogEntryElement struct {
	EntryName        string        `json:"entryName" yaml:"entryName"`
	EntryType        string        `json:"entryType" yaml:"entryType"`
	EntryDescription string        `json:"entryDescription" yaml:"entryDescription"`
	EntryMeta        rkentry.Entry `json:"entryMeta" yaml:"entryMeta"`
	OutputPaths      []string      `json:"outputPaths" yaml:"outputPaths"`
	ErrorOutputPaths []string      `json:"errorOutputPaths" yaml:"errorOutputPaths"`
}

// Helper function of /logs
func doLogsHelper(m map[string]rkentry.Entry, res *LogResponse) {
	entries := make([]*LogEntryElement, 0)

	// Iterate logger related entries and construct LogEntryElement
	for i := range m {
		entry := m[i]
		element := &LogEntryElement{
			EntryName:        entry.GetName(),
			EntryType:        entry.GetType(),
			EntryDescription: entry.GetDescription(),
			EntryMeta:        entry,
		}

		if val, ok := entry.(*rkentry.ZapLoggerEntry); ok {
			if val.LoggerConfig != nil {
				element.OutputPaths = val.LoggerConfig.OutputPaths
				element.ErrorOutputPaths = val.LoggerConfig.ErrorOutputPaths
			}
		}

		if val, ok := entry.(*rkentry.EventLoggerEntry); ok {
			if val.LoggerConfig != nil {
				element.OutputPaths = val.LoggerConfig.OutputPaths
				element.ErrorOutputPaths = val.LoggerConfig.ErrorOutputPaths
			}
		}

		entries = append(entries, element)
	}

	var entryType string

	if len(entries) > 0 {
		entryType = entries[0].EntryType
	}

	res.Entries[entryType] = entries
}

// Helper function of /logs
func doLogs(ctx *gin.Context) *LogResponse {
	res := &LogResponse{
		Entries: make(map[string][]*LogEntryElement),
	}

	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	if ctx == nil {
		return res
	}

	doLogsHelper(rkentry.GlobalAppCtx.ListEventLoggerEntriesRaw(), res)
	doLogsHelper(rkentry.GlobalAppCtx.ListZapLoggerEntriesRaw(), res)

	return res
}

// @Summary List logger related entries
// @Id 10
// @version 1.0
// @produce application/json
// @Success 200 {array} LogResponse
// @Router /v1/rk/logs [get]
func (entry *CommonServiceEntry) Logs(ctx *gin.Context) {
	if ctx == nil {
		return
	}
	// Add auto generated request ID
	rkginctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, doLogs(ctx))
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

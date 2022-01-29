// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgin an implementation of rkentry.Entry which could be used start restful server with gin framework
package rkgin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/middleware/auth"
	"github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/rookie-ninja/rk-entry/middleware/csrf"
	"github.com/rookie-ninja/rk-entry/middleware/jwt"
	"github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-entry/middleware/meta"
	"github.com/rookie-ninja/rk-entry/middleware/metrics"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/rookie-ninja/rk-entry/middleware/secure"
	"github.com/rookie-ninja/rk-entry/middleware/timeout"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/cors"
	"github.com/rookie-ninja/rk-gin/interceptor/csrf"
	"github.com/rookie-ninja/rk-gin/interceptor/gzip"
	"github.com/rookie-ninja/rk-gin/interceptor/jwt"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/meta"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	rkginpanic "github.com/rookie-ninja/rk-gin/interceptor/panic"
	"github.com/rookie-ninja/rk-gin/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-gin/interceptor/secure"
	"github.com/rookie-ninja/rk-gin/interceptor/timeout"
	"github.com/rookie-ninja/rk-gin/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
)

const (
	// GinEntryType type of entry
	GinEntryType = "Gin"
	// GinEntryDescription description of entry
	GinEntryDescription = "Internal RK entry which helps to bootstrap with Gin framework."
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap gin entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterGinEntriesWithConfig)
}

// BootConfig boot config which is for gin entry.
type BootConfig struct {
	Gin []struct {
		Enabled       bool                            `yaml:"enabled" json:"enabled"`
		Name          string                          `yaml:"name" json:"name"`
		Port          uint64                          `yaml:"port" json:"port"`
		Description   string                          `yaml:"description" json:"description"`
		CertEntry     string                          `yaml:"certEntry" json:"certEntry"`
		SW            rkentry.BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService rkentry.BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            rkentry.BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          rkentry.BootConfigProm          `yaml:"prom" json:"prom"`
		Static        rkentry.BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap  rkmidlog.BootConfig     `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm rkmidmetrics.BootConfig `yaml:"metricsProm" json:"metricsProm"`
			Auth        rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors        rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta        rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt         rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure      rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			RateLimit   rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Csrf        rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			Gzip        struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Level   string `yaml:"level" json:"level"`
			} `yaml:"gzip" json:"gzip"`
			Timeout          rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
			TracingTelemetry rkmidtrace.BootConfig   `yaml:"tracingTelemetry" json:"tracingTelemetry"`
		} `yaml:"interceptors" json:"interceptors"`
		Logger struct {
			ZapLogger   string `yaml:"zapLogger" json:"zapLogger"`
			EventLogger string `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"gin" json:"gin"`
}

// GinEntry implements rkentry.Entry interface.
type GinEntry struct {
	EntryName          string                          `json:"entryName" yaml:"entryName"`
	EntryType          string                          `json:"entryType" yaml:"entryType"`
	EntryDescription   string                          `json:"-" yaml:"-"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry         `json:"-" yaml:"-"`
	EventLoggerEntry   *rkentry.EventLoggerEntry       `json:"-" yaml:"-"`
	Router             *gin.Engine                     `json:"-" yaml:"-"`
	Server             *http.Server                    `json:"-" yaml:"-"`
	Port               uint64                          `json:"port" yaml:"port"`
	SwEntry            *rkentry.SwEntry                `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	TvEntry            *rkentry.TvEntry                `json:"-" yaml:"-"`
}

// RegisterGinEntriesWithConfig register gin entries with provided config file (Must YAML file).
//
// Currently, support two ways to provide config file path.
// 1: With function parameters
// 2: With command line flag "--rkboot" described in rkcommon.BootConfigPathFlagKey (Will override function parameter if exists)
// Command line flag has high priority which would override function parameter
//
// Error handling:
// Process will shutdown if any errors occur with rkcommon.ShutdownWithError function
//
// Override elements in config file:
// We learned from HELM source code which would override elements in YAML file with "--set" flag followed with comma
// separated key/value pairs.
//
// We are using "--rkset" described in rkcommon.BootConfigOverrideKey in order to distinguish with user flags
// Example of common usage: ./binary_file --rkset "key1=val1,key2=val2"
// Example of nested map:   ./binary_file --rkset "outer.inner.key=val"
// Example of slice:        ./binary_file --rkset "outer[0].key=val"
func RegisterGinEntriesWithConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init gin entries with boot config
	for i := range config.Gin {
		element := config.Gin[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		// Register swagger entry
		swEntry := rkentry.RegisterSwEntryWithConfig(&element.SW, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, element.CommonService.Enabled)

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntryWithConfig(&element.Prom, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, promRegistry)

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntryWithConfig(&element.CommonService, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register TV entry
		tvEntry := rkentry.RegisterTvEntryWithConfig(&element.TV, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntryWithConfig(&element.Static, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		inters := make([]gin.HandlerFunc, 0)

		// logging middlewares
		if element.Interceptors.LoggingZap.Enabled {
			inters = append(inters, rkginlog.Interceptor(
				rkmidlog.ToOptions(&element.Interceptors.LoggingZap, element.Name, GinEntryType,
					zapLoggerEntry, eventLoggerEntry)...))
		}

		// metrics middleware
		if element.Interceptors.MetricsProm.Enabled {
			inters = append(inters, rkginmetrics.Interceptor(
				rkmidmetrics.ToOptions(&element.Interceptors.MetricsProm, element.Name, GinEntryType,
					promRegistry, rkmidmetrics.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Interceptors.TracingTelemetry.Enabled {
			inters = append(inters, rkgintrace.Interceptor(
				rkmidtrace.ToOptions(&element.Interceptors.TracingTelemetry, element.Name, GinEntryType)...))
		}

		// jwt middleware
		if element.Interceptors.Jwt.Enabled {
			inters = append(inters, rkginjwt.Interceptor(
				rkmidjwt.ToOptions(&element.Interceptors.Jwt, element.Name, GinEntryType)...))
		}

		// secure middleware
		if element.Interceptors.Secure.Enabled {
			inters = append(inters, rkginsec.Interceptor(
				rkmidsec.ToOptions(&element.Interceptors.Secure, element.Name, GinEntryType)...))
		}

		// csrf middleware
		if element.Interceptors.Csrf.Enabled {
			inters = append(inters, rkgincsrf.Interceptor(
				rkmidcsrf.ToOptions(&element.Interceptors.Csrf, element.Name, GinEntryType)...))
		}

		// cors middleware
		if element.Interceptors.Cors.Enabled {
			inters = append(inters, rkgincors.Interceptor(
				rkmidcors.ToOptions(&element.Interceptors.Cors, element.Name, GinEntryType)...))
		}

		// gzip middleware
		if element.Interceptors.Gzip.Enabled {
			opts := []rkgingzip.Option{
				rkgingzip.WithEntryNameAndType(element.Name, GinEntryType),
				rkgingzip.WithLevel(element.Interceptors.Gzip.Level),
			}

			inters = append(inters, rkgingzip.Interceptor(opts...))
		}

		// meta middleware
		if element.Interceptors.Meta.Enabled {
			inters = append(inters, rkginmeta.Interceptor(
				rkmidmeta.ToOptions(&element.Interceptors.Meta, element.Name, GinEntryType)...))
		}

		// auth middlewares
		if element.Interceptors.Auth.Enabled {
			inters = append(inters, rkginauth.Interceptor(
				rkmidauth.ToOptions(&element.Interceptors.Auth, element.Name, GinEntryType)...))
		}

		// timeout middlewares
		if element.Interceptors.Timeout.Enabled {
			inters = append(inters, rkgintimeout.Interceptor(
				rkmidtimeout.ToOptions(&element.Interceptors.Timeout, element.Name, GinEntryType)...))
		}

		// rate limit middleware
		if element.Interceptors.RateLimit.Enabled {
			inters = append(inters, rkginlimit.Interceptor(
				rkmidlimit.ToOptions(&element.Interceptors.RateLimit, element.Name, GinEntryType)...))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.CertEntry)

		entry := RegisterGinEntry(
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry),
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithSwEntry(swEntry),
			WithPromEntry(promEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithCertEntry(certEntry),
			WithTvEntry(tvEntry),
			WithStaticFileHandlerEntry(staticEntry))

		entry.AddInterceptor(inters...)

		res[name] = entry
	}

	return res
}

// RegisterGinEntry register GinEntry with options.
func RegisterGinEntry(opts ...GinEntryOption) *GinEntry {
	entry := &GinEntry{
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		EntryType:        GinEntryType,
		EntryDescription: GinEntryDescription,
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "gin-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Router == nil {
		gin.SetMode(gin.ReleaseMode)
		entry.Router = gin.New()
	}

	if entry.Port != 0 {
		entry.Server = &http.Server{
			Addr:    "0.0.0.0:" + strconv.FormatUint(entry.Port, 10),
			Handler: entry.Router,
		}
	}

	// add entry name and entry type into loki syncer if enabled
	entry.ZapLoggerEntry.AddEntryLabelToLokiSyncer(entry)
	entry.EventLoggerEntry.AddEntryLabelToLokiSyncer(entry)

	// Default interceptor should be at front
	// insert panic interceptor
	entry.Router.Use(rkginpanic.Interceptor(
		rkmidpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *GinEntry) GetName() string {
	return entry.EntryName
}

// GetType Get entry type.
func (entry *GinEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry.
func (entry *GinEntry) GetDescription() string {
	return entry.EntryDescription
}

// Bootstrap GinEntry.
func (entry *GinEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap")

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		entry.Router.GET(path.Join(entry.SwEntry.Path, "*any"), gin.WrapF(entry.SwEntry.ConfigFileHandler()))
		entry.Router.GET(path.Join(entry.SwEntry.AssetsFilePath, "*any"), gin.WrapF(entry.SwEntry.AssetsFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		entry.Router.GET(path.Join(entry.StaticFileEntry.Path, "*any"), gin.WrapF(entry.StaticFileEntry.GetFileHandler()))
		entry.StaticFileEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.Router.GET(entry.PromEntry.Path, gin.WrapH(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Router.GET(entry.CommonServiceEntry.HealthyPath, gin.WrapF(entry.CommonServiceEntry.Healthy))
		entry.Router.GET(entry.CommonServiceEntry.GcPath, gin.WrapF(entry.CommonServiceEntry.Gc))
		entry.Router.GET(entry.CommonServiceEntry.InfoPath, gin.WrapF(entry.CommonServiceEntry.Info))
		entry.Router.GET(entry.CommonServiceEntry.ConfigsPath, gin.WrapF(entry.CommonServiceEntry.Configs))
		entry.Router.GET(entry.CommonServiceEntry.SysPath, gin.WrapF(entry.CommonServiceEntry.Sys))
		entry.Router.GET(entry.CommonServiceEntry.EntriesPath, gin.WrapF(entry.CommonServiceEntry.Entries))
		entry.Router.GET(entry.CommonServiceEntry.CertsPath, gin.WrapF(entry.CommonServiceEntry.Certs))
		entry.Router.GET(entry.CommonServiceEntry.LogsPath, gin.WrapF(entry.CommonServiceEntry.Logs))
		entry.Router.GET(entry.CommonServiceEntry.DepsPath, gin.WrapF(entry.CommonServiceEntry.Deps))
		entry.Router.GET(entry.CommonServiceEntry.LicensePath, gin.WrapF(entry.CommonServiceEntry.License))
		entry.Router.GET(entry.CommonServiceEntry.ReadmePath, gin.WrapF(entry.CommonServiceEntry.Readme))
		entry.Router.GET(entry.CommonServiceEntry.GitPath, gin.WrapF(entry.CommonServiceEntry.Git))

		// swagger doc already generated at rkentry.CommonService
		// follow bellow actions
		entry.Router.GET(entry.CommonServiceEntry.ApisPath, entry.Apis)
		entry.Router.GET(entry.CommonServiceEntry.ReqPath, entry.Req)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Router.RouterGroup.GET(path.Join(entry.TvEntry.BasePath, "*item"), entry.TV)
		entry.Router.GET(path.Join(entry.TvEntry.AssetsFilePath, "*any"), gin.WrapF(entry.TvEntry.AssetsFileHandler()))

		entry.TvEntry.Bootstrap(ctx)
	}

	// Start gin server
	go entry.startServer(event, logger)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt GinEntry.
func (entry *GinEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt")

	if entry.IsStaticFileHandlerEnabled() {
		// Interrupt entry
		entry.StaticFileEntry.Interrupt(ctx)
	}

	if entry.IsSwEnabled() {
		// Interrupt swagger entry
		entry.SwEntry.Interrupt(ctx)
	}

	if entry.IsPromEnabled() {
		// Interrupt prometheus entry
		entry.PromEntry.Interrupt(ctx)
	}

	if entry.IsCommonServiceEnabled() {
		// Interrupt common service entry
		entry.CommonServiceEntry.Interrupt(ctx)
	}

	if entry.IsTvEnabled() {
		// Interrupt common service entry
		entry.TvEntry.Interrupt(ctx)
	}

	if entry.Router != nil && entry.Server != nil {
		if err := entry.Server.Shutdown(context.Background()); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gin-server.", event.ListPayloads()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)

	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())
}

// String Stringfy gin entry.
func (entry *GinEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// ***************** Stringfy *****************

// MarshalJSON Marshal entry.
func (entry *GinEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":          entry.EntryName,
		"entryType":          entry.EntryType,
		"entryDescription":   entry.EntryDescription,
		"eventLoggerEntry":   entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":     entry.ZapLoggerEntry.GetName(),
		"port":               entry.Port,
		"swEntry":            entry.SwEntry,
		"commonServiceEntry": entry.CommonServiceEntry,
		"promEntry":          entry.PromEntry,
		"tvEntry":            entry.TvEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *GinEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ***************** Public functions *****************

// GetGinEntry Get GinEntry from rkentry.GlobalAppCtx.
func GetGinEntry(name string) *GinEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GinEntry)
	return entry
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *GinEntry) AddInterceptor(inters ...gin.HandlerFunc) {
	entry.Router.Use(inters...)
}

// IsSwEnabled Is swagger entry enabled?
func (entry *GinEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsStaticFileHandlerEnabled Is static file handler entry enabled?
func (entry *GinEntry) IsStaticFileHandlerEnabled() bool {
	return entry.StaticFileEntry != nil
}

// IsPromEnabled Is prometheus entry enabled?
func (entry *GinEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *GinEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsTvEnabled Is TV entry enabled?
func (entry *GinEntry) IsTvEnabled() bool {
	return entry.TvEntry != nil
}

// IsTlsEnabled Is TLS enabled?
func (entry *GinEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// ***************** Helper function *****************

// Add basic fields into event.
func (entry *GinEntry) logBasicInfo(operation string) (rkquery.Event, *zap.Logger) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	logger := entry.ZapLoggerEntry.GetLogger().With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.EntryName))

	// add general info
	event.AddPayloads(
		zap.Uint64("ginPort", entry.Port))

	// add SwEntry info
	if entry.IsSwEnabled() {
		event.AddPayloads(
			zap.Bool("swEnabled", true),
			zap.String("swPath", entry.SwEntry.Path))
	}

	// add CommonServiceEntry info
	if entry.IsCommonServiceEnabled() {
		event.AddPayloads(
			zap.Bool("commonServiceEnabled", true),
			zap.String("commonServicePathPrefix", "/rk/v1/"))
	}

	// add TvEntry info
	if entry.IsTvEnabled() {
		event.AddPayloads(
			zap.Bool("tvEnabled", true),
			zap.String("tvPath", "/rk/v1/tv/"))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.PromEntry.Port),
			zap.String("promPath", entry.PromEntry.Path))
	}

	// add StaticFileHandlerEntry info
	if entry.IsStaticFileHandlerEnabled() {
		event.AddPayloads(
			zap.Bool("staticFileHandlerEnabled", true),
			zap.String("staticFileHandlerPath", entry.StaticFileEntry.Path))
	}

	// add tls info
	if entry.IsTlsEnabled() {
		event.AddPayloads(
			zap.Bool("tlsEnabled", true))
	}

	logger.Info(fmt.Sprintf("%s ginEntry", operation))

	return event, logger
}

// Start server
// We move the code here for testability
func (entry *GinEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.Server != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			if cert, err := tls.X509KeyPair(entry.CertEntry.Store.ServerCert, entry.CertEntry.Store.ServerKey); err != nil {
				logger.Error("Error occurs while parsing TLS.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			} else {
				entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
			}

			if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while serving gin-listener-tls.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while serving gin-listener.", event.ListPayloads()...)
				entry.EventLoggerEntry.GetEventHelper().FinishWithCond(event, false)
				rkcommon.ShutdownWithError(err)
			}
		}
	}
}

// ***************** Common Service Extension API *****************

// Apis list apis from gin.Router
func (entry *GinEntry) Apis(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	ctx.JSON(http.StatusOK, entry.doApis(ctx))
}

// Req handler
func (entry *GinEntry) Req(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, entry.doReq(ctx))
}

// TV handler
func (entry *GinEntry) TV(ctx *gin.Context) {
	logger := rkginctx.GetLogger(ctx)

	contentType := "text/html; charset=utf-8"

	switch item := ctx.Param("item"); item {
	case "/apis":
		buf := entry.TvEntry.ExecuteTemplate("apis", entry.doApis(ctx), logger)
		ctx.Data(http.StatusOK, contentType, buf.Bytes())
	default:
		buf := entry.TvEntry.Action(item, logger)
		ctx.Data(http.StatusOK, contentType, buf.Bytes())
	}
}

// Helper function for APIs call
func (entry *GinEntry) doApis(ctx *gin.Context) *rkentry.ApisResponse {
	res := &rkentry.ApisResponse{
		Entries: make([]*rkentry.ApisResponseElement, 0),
	}

	routes := entry.Router.Routes()
	for j := range routes {
		info := routes[j]

		entry := &rkentry.ApisResponseElement{
			EntryName: entry.GetName(),
			Method:    info.Method,
			Path:      info.Path,
			Port:      entry.Port,
			SwUrl:     entry.constructSwUrl(ctx),
		}
		res.Entries = append(res.Entries, entry)
	}
	return res
}

// Construct swagger URL based on IP and scheme
func (entry *GinEntry) constructSwUrl(ctx *gin.Context) string {
	if !entry.IsSwEnabled() {
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

// Helper function for Req call
func (entry *GinEntry) doReq(ctx *gin.Context) *rkentry.ReqResponse {
	metricsSet := rkmidmetrics.GetServerMetricsSet(entry.GetName())
	if metricsSet == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	vector := metricsSet.GetSummary(rkmidmetrics.MetricsNameElapsedNano)
	if vector == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// Fill missed metrics
	apis := make([]string, 0)

	routes := entry.Router.Routes()
	for j := range routes {
		info := routes[j]
		apis = append(apis, info.Path)
	}

	// Add empty metrics into result
	for i := range apis {
		if !entry.containsMetrics(apis[i], reqMetrics) {
			reqMetrics = append(reqMetrics, &rkentry.ReqMetricsRK{
				RestPath: apis[i],
				ResCode:  make([]*rkentry.ResCodeRK, 0),
			})
		}
	}

	return &rkentry.ReqResponse{
		Metrics: reqMetrics,
	}
}

// Is metrics from prometheus contains particular api?
func (entry *GinEntry) containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].RestPath == api {
			return true
		}
	}

	return false
}

// ***************** Options *****************

// GinEntryOption Gin entry option.
type GinEntryOption func(*GinEntry)

// WithZapLoggerEntry provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntry(zapLogger *rkentry.ZapLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		if zapLogger != nil {
			entry.ZapLoggerEntry = zapLogger
		}
	}
}

// WithEventLoggerEntry provide rkentry.EventLoggerEntry.
func WithEventLoggerEntry(eventLogger *rkentry.EventLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		if eventLogger != nil {
			entry.EventLoggerEntry = eventLogger
		}
	}
}

// WithCommonServiceEntry provide CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithTvEntry provide TvEntry.
func WithTvEntry(tvEntry *rkentry.TvEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.TvEntry = tvEntry
	}
}

// WithStaticFileHandlerEntry provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntry(staticEntry *rkentry.StaticFileHandlerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SwEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.SwEntry = sw
	}
}

// WithPort provide port.
func WithPort(port uint64) GinEntryOption {
	return func(entry *GinEntry) {
		entry.Port = port
	}
}

// WithName provide name.
func WithName(name string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EntryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EntryDescription = description
	}
}

// WithPromEntry provide PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.PromEntry = prom
	}
}

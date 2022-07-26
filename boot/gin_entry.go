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
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkerror "github.com/rookie-ninja/rk-entry/v2/error"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	rkmidauth "github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	rkmidcors "github.com/rookie-ninja/rk-entry/v2/middleware/cors"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	rkmidjwt "github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	rkmidlog "github.com/rookie-ninja/rk-entry/v2/middleware/log"
	rkmidmeta "github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	rkmidpanic "github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	rkmidprom "github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	rkmidlimit "github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
	rkmidsec "github.com/rookie-ninja/rk-entry/v2/middleware/secure"
	rkmidtimeout "github.com/rookie-ninja/rk-entry/v2/middleware/timeout"
	rkmidtrace "github.com/rookie-ninja/rk-entry/v2/middleware/tracing"
	"github.com/rookie-ninja/rk-gin/v2/middleware/auth"
	rkgincors "github.com/rookie-ninja/rk-gin/v2/middleware/cors"
	"github.com/rookie-ninja/rk-gin/v2/middleware/csrf"
	"github.com/rookie-ninja/rk-gin/v2/middleware/gzip"
	"github.com/rookie-ninja/rk-gin/v2/middleware/jwt"
	"github.com/rookie-ninja/rk-gin/v2/middleware/log"
	"github.com/rookie-ninja/rk-gin/v2/middleware/meta"
	"github.com/rookie-ninja/rk-gin/v2/middleware/panic"
	"github.com/rookie-ninja/rk-gin/v2/middleware/prom"
	"github.com/rookie-ninja/rk-gin/v2/middleware/ratelimit"
	"github.com/rookie-ninja/rk-gin/v2/middleware/secure"
	"github.com/rookie-ninja/rk-gin/v2/middleware/timeout"
	"github.com/rookie-ninja/rk-gin/v2/middleware/tracing"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// GinEntryType type of entry
	GinEntryType = "GinEntry"
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap gin entry automatically from boot config file
func init() {
	rkentry.RegisterWebFrameRegFunc(RegisterGinEntryYAML)
}

// BootGin boot config which is for gin entry.
type BootGin struct {
	Gin []*BootGinElement `yaml:"gin" json:"gin"`
}

type BootGinElement struct {
	Enabled       bool                          `yaml:"enabled" json:"enabled"`
	Name          string                        `yaml:"name" json:"name"`
	Port          uint64                        `yaml:"port" json:"port"`
	Description   string                        `yaml:"description" json:"description"`
	SW            rkentry.BootSW                `yaml:"sw" json:"sw"`
	Docs          rkentry.BootDocs              `yaml:"docs" json:"docs"`
	CommonService rkentry.BootCommonService     `yaml:"commonService" json:"commonService"`
	Prom          rkentry.BootProm              `yaml:"prom" json:"prom"`
	CertEntry     string                        `yaml:"certEntry" json:"certEntry"`
	LoggerEntry   string                        `yaml:"loggerEntry" json:"loggerEntry"`
	EventEntry    string                        `yaml:"eventEntry" json:"eventEntry"`
	Static        rkentry.BootStaticFileHandler `yaml:"static" json:"static"`
	PProf         rkentry.BootPProf             `yaml:"pprof" json:"pprof"`
	Middleware    struct {
		Ignore     []string                `yaml:"ignore" json:"ignore"`
		ErrorModel string                  `yaml:"errorModel" json:"errorModel"`
		Logging    rkmidlog.BootConfig     `yaml:"logging" json:"logging"`
		Prom       rkmidprom.BootConfig    `yaml:"prom" json:"prom"`
		Auth       rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
		Cors       rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
		Meta       rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
		Jwt        rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
		Secure     rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
		RateLimit  rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
		Csrf       rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
		Timeout    rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
		Trace      rkmidtrace.BootConfig   `yaml:"trace" json:"trace"`
		Gzip       struct {
			Enabled bool     `yaml:"enabled" json:"enabled"`
			Ignore  []string `yaml:"ignore" json:"ignore"`
			Level   string   `yaml:"level" json:"level"`
		} `yaml:"gzip" json:"gzip"`
	} `yaml:"middleware" json:"middleware"`
}

// GinEntry implements rkentry.Entry interface.
type GinEntry struct {
	entryName          string                          `json:"-" yaml:"-"`
	entryType          string                          `json:"-" yaml:"-"`
	entryDescription   string                          `json:"-" yaml:"-"`
	Router             *gin.Engine                     `json:"-" yaml:"-"`
	Server             *http.Server                    `json:"-" yaml:"-"`
	Port               uint64                          `json:"-" yaml:"-"`
	LoggerEntry        *rkentry.LoggerEntry            `json:"-" yaml:"-"`
	EventEntry         *rkentry.EventEntry             `json:"-" yaml:"-"`
	SwEntry            *rkentry.SWEntry                `json:"-" yaml:"-"`
	DocsEntry          *rkentry.DocsEntry              `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	PProfEntry         *rkentry.PProfEntry             `json:"-" yaml:"-"`
	bootstrapLogOnce   sync.Once                       `json:"-" yaml:"-"`
}

// RegisterGinEntryYAML register gin entries with provided config file (Must YAML file).
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
func RegisterGinEntryYAML(raw []byte) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootGin{}
	rkentry.UnmarshalBootYAML(raw, config)

	// 2: Init gin entries with boot config
	for i := range config.Gin {
		element := config.Gin[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		// logger entry
		loggerEntry := rkentry.GlobalAppCtx.GetLoggerEntry(element.LoggerEntry)
		if loggerEntry == nil {
			loggerEntry = rkentry.LoggerEntryStdout
		}

		// event entry
		eventEntry := rkentry.GlobalAppCtx.GetEventEntry(element.EventEntry)
		if eventEntry == nil {
			eventEntry = rkentry.EventEntryStdout
		}

		// cert entry
		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.CertEntry)

		// Register swagger entry
		swEntry := rkentry.RegisterSWEntry(&element.SW, rkentry.WithNameSWEntry(element.Name))

		// Register docs entry
		docsEntry := rkentry.RegisterDocsEntry(&element.Docs, rkentry.WithNameDocsEntry(element.Name))

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntry(&element.Prom, rkentry.WithRegistryPromEntry(promRegistry))

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntry(&element.CommonService)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntry(&element.Static, rkentry.WithNameStaticFileHandlerEntry(element.Name))

		// Register pprof entry
		pprofEntry := rkentry.RegisterPProfEntry(&element.PProf, rkentry.WithNamePProfEntry(element.Name))

		inters := make([]gin.HandlerFunc, 0)

		// add global path ignorance
		rkmid.AddPathToIgnoreGlobal(element.Middleware.Ignore...)

		// set error builder based on error builder
		switch strings.ToLower(element.Middleware.ErrorModel) {
		case "", "google":
			rkmid.SetErrorBuilder(rkerror.NewErrorBuilderGoogle())
		case "amazon":
			rkmid.SetErrorBuilder(rkerror.NewErrorBuilderAMZN())
		}

		// logging middlewares
		if element.Middleware.Logging.Enabled {
			inters = append(inters, rkginlog.Middleware(
				rkmidlog.ToOptions(&element.Middleware.Logging, element.Name, GinEntryType,
					loggerEntry, eventEntry)...))
		}

		// Default interceptor should be placed after logging middleware, we should make sure interceptors never panic
		// insert panic interceptor
		inters = append(inters, rkginpanic.Middleware(
			rkmidpanic.WithEntryNameAndType(element.Name, GinEntryType)))

		// metrics middleware
		if element.Middleware.Prom.Enabled {
			inters = append(inters, rkginprom.Middleware(
				rkmidprom.ToOptions(&element.Middleware.Prom, element.Name, GinEntryType,
					promRegistry, rkmidprom.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Middleware.Trace.Enabled {
			inters = append(inters, rkgintrace.Middleware(
				rkmidtrace.ToOptions(&element.Middleware.Trace, element.Name, GinEntryType)...))
		}

		// cors middleware
		if element.Middleware.Cors.Enabled {
			inters = append(inters, rkgincors.Middleware(
				rkmidcors.ToOptions(&element.Middleware.Cors, element.Name, GinEntryType)...))
		}

		// jwt middleware
		if element.Middleware.Jwt.Enabled {
			inters = append(inters, rkginjwt.Middleware(
				rkmidjwt.ToOptions(&element.Middleware.Jwt, element.Name, GinEntryType)...))
		}

		// secure middleware
		if element.Middleware.Secure.Enabled {
			inters = append(inters, rkginsec.Middleware(
				rkmidsec.ToOptions(&element.Middleware.Secure, element.Name, GinEntryType)...))
		}

		// csrf middleware
		if element.Middleware.Csrf.Enabled {
			inters = append(inters, rkgincsrf.Middleware(
				rkmidcsrf.ToOptions(&element.Middleware.Csrf, element.Name, GinEntryType)...))
		}

		// gzip middleware
		if element.Middleware.Gzip.Enabled {
			opts := []rkgingzip.Option{
				rkgingzip.WithEntryNameAndType(element.Name, GinEntryType),
				rkgingzip.WithLevel(element.Middleware.Gzip.Level),
				rkgingzip.WithPathToIgnore(element.Middleware.Gzip.Ignore...),
			}

			inters = append(inters, rkgingzip.Middleware(opts...))
		}

		// meta middleware
		if element.Middleware.Meta.Enabled {
			inters = append(inters, rkginmeta.Middleware(
				rkmidmeta.ToOptions(&element.Middleware.Meta, element.Name, GinEntryType)...))
		}

		// auth middlewares
		if element.Middleware.Auth.Enabled {
			inters = append(inters, rkginauth.Middleware(
				rkmidauth.ToOptions(&element.Middleware.Auth, element.Name, GinEntryType)...))
		}

		// timeout middlewares
		if element.Middleware.Timeout.Enabled {
			inters = append(inters, rkgintout.Middleware(
				rkmidtimeout.ToOptions(&element.Middleware.Timeout, element.Name, GinEntryType)...))
		}

		// rate limit middleware
		if element.Middleware.RateLimit.Enabled {
			inters = append(inters, rkginlimit.Middleware(
				rkmidlimit.ToOptions(&element.Middleware.RateLimit, element.Name, GinEntryType)...))
		}

		entry := RegisterGinEntry(
			WithLoggerEntry(loggerEntry),
			WithEventEntry(eventEntry),
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithSwEntry(swEntry),
			WithDocsEntry(docsEntry),
			WithPromEntry(promEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithCertEntry(certEntry),
			WithPProfEntry(pprofEntry),
			WithStaticFileHandlerEntry(staticEntry))

		entry.AddMiddleware(inters...)

		res[name] = entry
	}

	return res
}

// RegisterGinEntry register GinEntry with options.
func RegisterGinEntry(opts ...GinEntryOption) *GinEntry {
	entry := &GinEntry{
		entryType:        GinEntryType,
		entryDescription: "Internal RK entry which helps to bootstrap with Gin framework.",
		LoggerEntry:      rkentry.NewLoggerEntryStdout(),
		EventEntry:       rkentry.NewEventEntryStdout(),
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "gin-" + strconv.FormatUint(entry.Port, 10)
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
	entry.LoggerEntry.AddEntryLabelToLokiSyncer(entry)
	entry.EventEntry.AddEntryLabelToLokiSyncer(entry)

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *GinEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *GinEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *GinEntry) GetDescription() string {
	return entry.entryDescription
}

// Bootstrap GinEntry.
func (entry *GinEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap", ctx)

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Router.GET(entry.CommonServiceEntry.ReadyPath, gin.WrapF(entry.CommonServiceEntry.Ready))
		entry.Router.GET(entry.CommonServiceEntry.AlivePath, gin.WrapF(entry.CommonServiceEntry.Alive))
		entry.Router.GET(entry.CommonServiceEntry.GcPath, gin.WrapF(entry.CommonServiceEntry.Gc))
		entry.Router.GET(entry.CommonServiceEntry.InfoPath, gin.WrapF(entry.CommonServiceEntry.Info))

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		entry.Router.GET(path.Join(entry.SwEntry.Path, "*any"), gin.WrapF(entry.SwEntry.ConfigFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is docs enabled?
	if entry.IsDocsEnabled() {
		entry.Router.GET(path.Join(entry.DocsEntry.Path, "*any"), gin.WrapF(entry.DocsEntry.ConfigFileHandler()))
		entry.DocsEntry.Bootstrap(ctx)
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

	// Is pprof enabled?
	if entry.IsPProfEnabled() {
		pprof.Register(entry.Router, entry.PProfEntry.Path)
	}

	// Start gin server
	go entry.startServer(event, logger)

	entry.bootstrapLogOnce.Do(func() {
		// Print link and logging message
		scheme := "http"
		if entry.IsTlsEnabled() {
			scheme = "https"
		}

		if entry.IsSwEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("SwaggerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.SwEntry.Path))
		}
		if entry.IsDocsEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("DocsEntry: %s://localhost:%d%s", scheme, entry.Port, entry.DocsEntry.Path))
		}
		if entry.IsPromEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PromEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PromEntry.Path))
		}
		if entry.IsStaticFileHandlerEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("StaticFileHandlerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.StaticFileEntry.Path))
		}
		if entry.IsCommonServiceEnabled() {
			handlers := []string{
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.ReadyPath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.AlivePath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.InfoPath),
			}

			entry.LoggerEntry.Info(fmt.Sprintf("CommonSreviceEntry: %s", strings.Join(handlers, ", ")))
		}
		if entry.IsPProfEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PProfEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PProfEntry.Path))
		}
		entry.EventEntry.Finish(event)
	})
}

// Interrupt GinEntry.
func (entry *GinEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt", ctx)

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

	if entry.IsDocsEnabled() {
		entry.DocsEntry.Interrupt(ctx)
	}

	if entry.IsPProfEnabled() {
		entry.PProfEntry.Interrupt(ctx)
	}

	if entry.Router != nil && entry.Server != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := entry.Server.Shutdown(ctx); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gin-server.", event.ListPayloads()...)
		}
	}

	entry.EventEntry.Finish(event)

	rkentry.GlobalAppCtx.RemoveEntry(entry)
}

// String Stringfy gin entry.
func (entry *GinEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// SetReadinessCheck set readiness check into rkentry.GlobalAppCtx
func (entry *GinEntry) SetReadinessCheck(f rkentry.ReadinessCheck) {
	rkentry.GlobalAppCtx.SetReadinessCheck(f)
}

// SetLivenessCheck set liveness check into rkentry.GlobalAppCtx
func (entry *GinEntry) SetLivenessCheck(f rkentry.LivenessCheck) {
	rkentry.GlobalAppCtx.SetLivenessCheck(f)
}

// ***************** Stringfy *****************

// MarshalJSON Marshal entry.
func (entry *GinEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":                   entry.entryName,
		"type":                   entry.entryType,
		"description":            entry.entryDescription,
		"port":                   entry.Port,
		"swEntry":                entry.SwEntry,
		"docsEntry":              entry.DocsEntry,
		"commonServiceEntry":     entry.CommonServiceEntry,
		"promEntry":              entry.PromEntry,
		"staticFileHandlerEntry": entry.StaticFileEntry,
		"pprofEntry":             entry.PProfEntry,
	}

	if entry.IsTlsEnabled() {
		m["certEntry"] = entry.CertEntry
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
	entryRaw := rkentry.GlobalAppCtx.GetEntry(GinEntryType, name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GinEntry)
	return entry
}

// AddMiddleware Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *GinEntry) AddMiddleware(mids ...gin.HandlerFunc) {
	entry.Router.Use(mids...)
}

// IsSwEnabled Is swagger entry enabled?
func (entry *GinEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsDocsEnabled Is docs entry enabled?
func (entry *GinEntry) IsDocsEnabled() bool {
	return entry.DocsEntry != nil
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

// IsPProfEnabled Is pprof entry enabled?
func (entry *GinEntry) IsPProfEnabled() bool {
	return entry.PProfEntry != nil
}

// IsTlsEnabled Is TLS enabled?
func (entry *GinEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Certificate != nil
}

// ***************** Helper function *****************

// Add basic fields into event.
func (entry *GinEntry) logBasicInfo(operation string, ctx context.Context) (rkquery.Event, *zap.Logger) {
	event := entry.EventEntry.Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))

	// extract eventId if exists
	if val := ctx.Value("eventId"); val != nil {
		if id, ok := val.(string); ok {
			event.SetEventId(id)
		}
	}

	logger := entry.LoggerEntry.With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.entryName),
		zap.String("entryType", entry.entryType))

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
			zap.Bool("commonServiceEnabled", true))
	}

	// add DocsEntry info
	if entry.IsDocsEnabled() {
		event.AddPayloads(
			zap.Bool("docsEnabled", true),
			zap.String("docsPath", entry.DocsEntry.Path))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.Port),
			zap.String("promPath", entry.PromEntry.Path))
	}

	// add StaticFileHandlerEntry info
	if entry.IsStaticFileHandlerEnabled() {
		event.AddPayloads(
			zap.Bool("staticFileHandlerEnabled", true),
			zap.String("staticFileHandlerPath", entry.StaticFileEntry.Path))
	}

	// add PProfEntry info
	if entry.IsPProfEnabled() {
		event.AddPayloads(
			zap.Bool("pprofEnabled", true),
			zap.String("pprofPath", entry.PProfEntry.Path))
	}

	// add tls info
	if entry.IsTlsEnabled() {
		event.AddPayloads(
			zap.Bool("tlsEnabled", true))
	}

	logger.Info(fmt.Sprintf("%s GinEntry", operation))

	return event, logger
}

// Start server
// We move the code here for testability
func (entry *GinEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.Server != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*entry.CertEntry.Certificate}}

			if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while serving gin-listener-tls.", event.ListPayloads()...)
				entry.bootstrapLogOnce.Do(func() {
					entry.EventEntry.FinishWithCond(event, false)
				})
				rkentry.ShutdownWithError(err)
			}
		} else {
			if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while serving gin-listener.", event.ListPayloads()...)
				entry.bootstrapLogOnce.Do(func() {
					entry.EventEntry.FinishWithCond(event, false)
				})
				rkentry.ShutdownWithError(err)
			}
		}
	}
}

// ***************** Options *****************

// GinEntryOption Gin entry option.
type GinEntryOption func(*GinEntry)

// WithLoggerEntry provide rkentry.LoggerEntry.
func WithLoggerEntry(logger *rkentry.LoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		if logger != nil {
			entry.LoggerEntry = logger
		}
	}
}

// WithEventEntry provide rkentry.EventLoggerEntry.
func WithEventEntry(eventLogger *rkentry.EventEntry) GinEntryOption {
	return func(entry *GinEntry) {
		if eventLogger != nil {
			entry.EventEntry = eventLogger
		}
	}
}

// WithCommonServiceEntry provide CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CommonServiceEntry = commonServiceEntry
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
func WithSwEntry(sw *rkentry.SWEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.SwEntry = sw
	}
}

// WithDocsEntry provide SwEntry.
func WithDocsEntry(docs *rkentry.DocsEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.DocsEntry = docs
	}
}

func WithPProfEntry(p *rkentry.PProfEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.PProfEntry = p
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
		entry.entryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.entryDescription = description
	}
}

// WithPromEntry provide PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.PromEntry = prom
	}
}

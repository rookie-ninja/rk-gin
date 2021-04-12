// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	"github.com/rookie-ninja/rk-prom"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap gin entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterGinEntriesWithConfig)
}

// Boot config which is for gin entry.
//
// 1: Gin.Name: Name of gin entry, should be unique globally.
// 2: Gin.Port: Port of gin entry.
// 3: Gin.Cert.Ref: Reference of rkentry.CertEntry.
// 4: Gin.SW: See BootConfigSW for details.
// 5: Gin.CommonService: See BootConfigCommonService for details.
// 6: Gin.TV: See BootConfigTV for details.
// 7: Gin.Prom: See BootConfigProm for details.
// 8: Gin.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 9: Gin.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 10: Gin.Interceptors.BasicAuth.Enabled: Enable basic auth.
// 11: Gin.interceptors.BasicAuth.Credentials: Credential for basic auth, scheme: <user:pass>
// 12: Gin.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 13: Gin.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
type bootConfig struct {
	Gin []struct {
		Name string `yaml:"name"`
		Port uint64 `yaml:"port"`
		Cert struct {
			Ref string `yaml:"ref"`
		} `yaml:"cert"`
		SW            BootConfigSW            `yaml:"sw"`
		CommonService BootConfigCommonService `yaml:"commonService"`
		TV            BootConfigTV            `yaml:"tv"`
		Prom          BootConfigProm          `yaml:"prom"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"metricsProm"`
			BasicAuth struct {
				Enabled     bool     `yaml:"enabled"`
				Credentials []string `yaml:"credentials"`
			} `yaml:"basicAuth"`
		} `yaml:"interceptors"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref"`
			} `yaml:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref"`
			} `yaml:"eventLogger"`
		} `yaml:"logger"`
	} `yaml:"gin"`
}

// GinEntry implements rkentry.Entry interface.
//
// 1: ZapLoggerEntry: See rkentry.ZapLoggerEntry for details.
// 2: EventLoggerEntry: See rkentry.EventLoggerEntry for details.
// 3: Router: gin.Engine created while bootstrapping.
// 4: Server: http.Server created while bootstrapping.
// 5: Port: http/https port server listen to.
// 6: Interceptors: Interceptors user enabled from YAML config, by default, rkginpanic.PanicInterceptor would be injected.
// 7: SWEntry: See SWEntry for details.
// 8: CertStore: Created while bootstrapping with CertEntry from YAML config.
// 9: CommonServiceEntry: See CommonServiceEntry for details.
// 10: PromEntry: See PromEntry for details.
// 11: TVEntry: See TVEntry for details.
type GinEntry struct {
	ZapLoggerEntry     *rkentry.ZapLoggerEntry
	EventLoggerEntry   *rkentry.EventLoggerEntry
	Router             *gin.Engine
	Server             *http.Server
	entryName          string
	entryType          string
	Port               uint64
	Interceptors       []gin.HandlerFunc
	SWEntry            *SWEntry
	CertStore          *rkentry.CertStore
	CommonServiceEntry *CommonServiceEntry
	PromEntry          *PromEntry
	TVEntry            *TVEntry
}

// Gin entry option.
type GinEntryOption func(*GinEntry)

// Get GinEntry from rkentry.GlobalAppCtx.
func GetGinEntry(name string) *GinEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GinEntry)
	return entry
}

// Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryGin(zapLogger *rkentry.ZapLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryGin(eventLogger *rkentry.EventLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// Provide user interceptors.
func WithInterceptorsGin(inters ...gin.HandlerFunc) GinEntryOption {
	return func(entry *GinEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]gin.HandlerFunc, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// Provide CommonServiceEntry.
func WithCommonServiceEntryGin(commonServiceEntry *CommonServiceEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// Provide TVEntry.
func WithTVEntryGin(tvEntry *TVEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.TVEntry = tvEntry
	}
}

// Provide CertStore.
func WithCertStoreGin(store *rkentry.CertStore) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CertStore = store
	}
}

// Provide SWEntry.
func WithSWEntryGin(sw *SWEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.SWEntry = sw
	}
}

// Provide port.
func WithPortGin(port uint64) GinEntryOption {
	return func(entry *GinEntry) {
		entry.Port = port
	}
}

// Provide name.
func WithNameGin(name string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.entryName = name
	}
}

// Provide PromEntry.
func WithPromEntryGin(prom *PromEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.PromEntry = prom
	}
}

// Register gin entries with provided config file (Must YAML file).
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

	// 1: decode config map into boot config struct
	config := &bootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 3: init gin entries with boot config
	for i := range config.Gin {
		element := config.Gin[i]
		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry()

		// did we enabled swagger?
		var swEntry *SWEntry
		if element.SW.Enabled {
			// init swagger custom headers from config
			headers := make(map[string]string, 0)
			for i := range element.SW.Headers {
				header := element.SW.Headers[i]
				tokens := strings.Split(header, ":")
				if len(tokens) == 2 {
					headers[tokens[0]] = tokens[1]
				}
			}

			swEntry = NewSWEntry(
				WithNameSW(fmt.Sprintf("gin-sw-%d", element.Port)),
				WithZapLoggerEntrySW(zapLoggerEntry),
				WithEventLoggerEntrySW(eventLoggerEntry),
				WithPortSW(element.Port),
				WithPathSW(element.SW.Path),
				WithJSONPathSW(element.SW.JSONPath),
				WithHeadersSW(headers))
		}

		// did we enabled prometheus?
		var promEntry *PromEntry
		if element.Prom.Enabled {
			var pusher *rkprom.PushGatewayPusher
			if element.Prom.Pusher.Enabled {
				var certStore *rkentry.CertStore

				if certEntry != nil {
					certStore = certEntry.Stores[element.Prom.Pusher.Cert.Ref]
				}

				pusher, _ = rkprom.NewPushGatewayPusher(
					rkprom.WithIntervalMSPusher(time.Duration(element.Prom.Pusher.IntervalMS)*time.Millisecond),
					rkprom.WithRemoteAddressPusher(element.Prom.Pusher.RemoteAddress),
					rkprom.WithJobNamePusher(element.Prom.Pusher.JobName),
					rkprom.WithBasicAuthPusher(element.Prom.Pusher.BasicAuth),
					rkprom.WithZapLoggerEntryPusher(zapLoggerEntry),
					rkprom.WithEventLoggerEntryPusher(eventLoggerEntry),
					rkprom.WithCertStorePusher(certStore))
			}

			var certStore *rkentry.CertStore
			if certEntry != nil {
				certStore = certEntry.Stores[element.Prom.Cert.Ref]
			}

			promEntry = NewPromEntry(
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithCertStoreProm(certStore),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithEventLoggerEntryProm(eventLoggerEntry),
				WithPusherProm(pusher))

			if promEntry.Pusher != nil {
				promEntry.Pusher.SetGatherer(promEntry.Gatherer)
			}
		}

		inters := make([]gin.HandlerFunc, 0)
		// did we enabled logging interceptor?
		if element.Interceptors.LoggingZap.Enabled {
			opts := []rkginlog.Option{
				rkginlog.WithEntryName(element.Name),
				rkginlog.WithEventFactory(eventLoggerEntry.GetEventFactory()),
				rkginlog.WithLogger(zapLoggerEntry.GetLogger()),
			}

			inters = append(inters, rkginlog.LoggingZapInterceptor(opts...))
		}

		// did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkginmetrics.Option{
				rkginmetrics.WithEntryName(element.Name),
			}

			if promEntry != nil {
				opts = append(opts, rkginmetrics.WithRegisterer(promEntry.Registerer))
			}

			inters = append(inters, rkginmetrics.MetricsPromInterceptor(opts...))
		}

		// did we enabled auth interceptor?
		if element.Interceptors.BasicAuth.Enabled {
			accounts := gin.Accounts{}
			for i := range element.Interceptors.BasicAuth.Credentials {
				cred := element.Interceptors.BasicAuth.Credentials[i]
				tokens := strings.Split(cred, ":")
				if len(tokens) == 2 {
					accounts[tokens[0]] = tokens[1]
				}
			}
			inters = append(inters, rkginauth.BasicAuthInterceptor(accounts, element.Name))
		}

		// did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("gin-common-service-%d", element.Port)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry),
				WithPathPrefixCommonService(element.CommonService.PathPrefix))
		}

		// did we enabled tv?
		var tvEntry *TVEntry
		if element.TV.Enabled {
			tvEntry = NewTVEntry(
				WithNameTV(fmt.Sprintf("gin-tv-%d", element.Port)),
				WithZapLoggerEntryTV(zapLoggerEntry),
				WithEventLoggerEntryTV(eventLoggerEntry),
				WithPathPrefixTV(element.TV.PathPrefix))
		}

		var certStore *rkentry.CertStore
		if certEntry != nil {
			certStore = certEntry.Stores[element.Cert.Ref]
		}

		entry := RegisterGinEntry(
			WithZapLoggerEntryGin(zapLoggerEntry),
			WithEventLoggerEntryGin(eventLoggerEntry),
			WithNameGin(name),
			WithPortGin(element.Port),
			WithSWEntryGin(swEntry),
			WithPromEntryGin(promEntry),
			WithCommonServiceEntryGin(commonServiceEntry),
			WithCertStoreGin(certStore),
			WithTVEntryGin(tvEntry),
			WithInterceptorsGin(inters...))

		res[name] = entry
	}

	return res
}

// Register GinEntry with options.
func RegisterGinEntry(opts ...GinEntryOption) *GinEntry {
	entry := &GinEntry{
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		entryType:        "gin",
		Port:             80,
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

	if len(entry.entryName) < 1 {
		entry.entryName = "gin-server-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Interceptors == nil {
		entry.Interceptors = make([]gin.HandlerFunc, 0)
	}

	if entry.Router == nil {
		gin.SetMode(gin.ReleaseMode)
		entry.Router = gin.New()
	}

	// default interceptor should be at front
	entry.Interceptors = append(entry.Interceptors, rkginpanic.PanicInterceptor())
	entry.Router.Use(entry.Interceptors...)

	// init server only if port is not zero
	if entry.Port != 0 {
		entry.Server = &http.Server{
			Addr:    "0.0.0.0:" + strconv.FormatUint(entry.Port, 10),
			Handler: entry.Router,
		}
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Get entry name.
func (entry *GinEntry) GetName() string {
	return entry.entryName
}

// Get entry type.
func (entry *GinEntry) GetType() string {
	return entry.entryType
}

// Stringfy gin entry.
func (entry *GinEntry) String() string {
	m := map[string]interface{}{
		"entry_name":             entry.entryName,
		"entry_type":             entry.entryType,
		"port":                   strconv.FormatUint(entry.Port, 10),
		"interceptors_count":     strconv.Itoa(len(entry.Interceptors)),
		"sw_enabled":             strconv.FormatBool(entry.IsSWEnabled()),
		"tls_enabled":            strconv.FormatBool(entry.IsTLSEnabled()),
		"common_service_enabled": strconv.FormatBool(entry.IsCommonServiceEnabled()),
		"tv_enabled":             strconv.FormatBool(entry.IsTVEnabled()),
	}

	if entry.IsSWEnabled() {
		m["sw"] = rkcommon.ConvertJSONToMap(entry.SWEntry.String())
	}

	if entry.IsCommonServiceEnabled() {
		m["common_service"] = rkcommon.ConvertJSONToMap(entry.CommonServiceEntry.String())
	}

	if entry.IsTVEnabled() {
		m["tv"] = rkcommon.ConvertJSONToMap(entry.TVEntry.String())
	}

	bytes, err := json.Marshal(m)

	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Warn("failed to marshal gin entry to string", zap.Error(err))
		return "{}"
	}

	return string(bytes)
}

// Register interceptor, please make sure call this function before Bootstrap().
func (entry *GinEntry) RegisterInterceptor(interceptor gin.HandlerFunc) {
	if interceptor == nil {
		return
	}

	entry.Interceptors = append(entry.Interceptors, interceptor)
}

// Bootstrap GinEntry.
func (entry *GinEntry) Bootstrap(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start("bootstrap")

	ctx = context.Background()

	fields := []zap.Field{
		zap.String("entry_type", "gin"),
		zap.Uint64("entry_port", entry.Port),
		zap.String("entry_name", entry.entryName),
	}

	// Is swagger enabled?
	if entry.IsSWEnabled() {
		fields = append(fields,
			zap.String("sw_path", entry.SWEntry.Path),
			zap.Bool("sw_enabled", true))

		// Register swagger path into Router.
		entry.Router.GET(path.Join(entry.SWEntry.Path, "*any"), entry.SWEntry.GinHandler())
		entry.Router.GET("/swagger/*any", entry.SWEntry.GinFileHandler())

		// Bootstrap swagger entry.
		entry.SWEntry.Bootstrap(ctx)
	} else {
		fields = append(fields, zap.Bool("sw_enabled", false))
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		fields = append(fields, zap.String("prom_path", entry.PromEntry.Path),
			zap.Uint64("prom_port", entry.PromEntry.Port),
			zap.Bool("prom_enabled", true))

		// Register prom path into Router.
		entry.Router.GET(entry.PromEntry.Path, gin.WrapH(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))

		// don't start with http handler, we will handle it by ourselves
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		fields = append(fields, zap.Bool("common_service_enabled", true))

		// Register common service path into Router.
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"healthy", entry.CommonServiceEntry.Healthy)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"gc", entry.CommonServiceEntry.GC)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"info", entry.CommonServiceEntry.Info)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"config", entry.CommonServiceEntry.Config)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"apis", entry.CommonServiceEntry.APIs)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"sys", entry.CommonServiceEntry.Sys)
		entry.Router.GET(entry.CommonServiceEntry.PathPrefix+"req", entry.CommonServiceEntry.Req)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	} else {
		fields = append(fields, zap.Bool("common_service_enabled", false))
	}

	// Is TV enabled?
	if entry.IsTVEnabled() {
		fields = append(fields, zap.Bool("tv_enabled", true))

		// Bootstrap TV entry.
		entry.TVEntry.Bootstrap(ctx)
	} else {
		fields = append(fields, zap.Bool("tv_enabled", false))
	}

	// Is TLS enabled?
	if entry.IsTLSEnabled() {
		fields = append(fields, zap.Bool("tls_enabled", true))
	} else {
		fields = append(fields, zap.Bool("tls_enabled", false))
	}

	event.AddFields(fields...)

	// Start gin server
	if entry.Server != nil {
		entry.ZapLoggerEntry.GetLogger().Info("starting gin-server", fields...)
		entry.EventLoggerEntry.GetEventHelper().Finish(event)

		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTLSEnabled() {
			if cert, err := tls.X509KeyPair(entry.CertStore.ServerCert, entry.CertStore.ServerKey); err != nil {
				rkcommon.ShutdownWithError(err)
			} else {
				entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
			}

			if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				fields := append(fields, zap.Error(err))
				entry.ZapLoggerEntry.GetLogger().Error("error while serving gin-listener-tls", fields...)
				entry.EventLoggerEntry.GetEventHelper().FinishWithCond(event, false)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fields := append(fields, zap.Error(err))
				entry.ZapLoggerEntry.GetLogger().Error("error while serving gin-listener", fields...)
				entry.EventLoggerEntry.GetEventHelper().FinishWithCond(event, false)
				rkcommon.ShutdownWithError(err)
			}
		}
	}
}

func (entry *GinEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start("interrupt")

	fields := []zap.Field{
		zap.String("entry_type", "gin"),
		zap.Uint64("entry_port", entry.Port),
		zap.String("entry_name", entry.entryName),
	}

	if entry.IsTLSEnabled() {
		fields = append(fields, zap.Bool("tls_enabled", true))
	} else {
		fields = append(fields, zap.Bool("tls_enabled", false))
	}

	if entry.IsSWEnabled() {
		fields = append(fields, zap.String("sw_path", entry.SWEntry.Path))
		fields = append(fields, zap.Bool("sw_enabled", true))

		// Interrupt swagger entry
		entry.SWEntry.Interrupt(ctx)
	} else {
		fields = append(fields, zap.Bool("sw_enabled", false))
	}

	if entry.IsPromEnabled() {
		fields = append(fields, zap.String("prom_path", entry.PromEntry.Path),
			zap.Uint64("prom_port", entry.PromEntry.Port),
			zap.Bool("prom_enabled", true))

		// Interrupt prometheus entry
		entry.PromEntry.Interrupt(ctx)
	} else {
		fields = append(fields, zap.Bool("prom_enabled", false))
	}

	if entry.IsCommonServiceEnabled() {
		fields = append(fields, zap.Bool("common_service_enabled", true))

		// Interrupt common service entry
		entry.CommonServiceEntry.Interrupt(ctx)
	} else {
		fields = append(fields, zap.Bool("common_service_enabled", false))
	}

	if entry.IsTVEnabled() {
		fields = append(fields, zap.Bool("tv_enabled", true))

		// Interrupt common service entry
		entry.TVEntry.Interrupt(ctx)
	} else {
		fields = append(fields, zap.Bool("tv_enabled", false))
	}

	if entry.Router != nil && entry.Server != nil {
		entry.ZapLoggerEntry.GetLogger().Info("stopping gin-server", fields...)
		if err := entry.Server.Shutdown(context.Background()); err != nil {
			fields = append(fields, zap.Error(err))
			entry.ZapLoggerEntry.GetLogger().Warn("error occurs while stopping gin-server", fields...)
		}
	}

	event.AddFields(fields...)
	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Is swagger entry enabled?
func (entry *GinEntry) IsSWEnabled() bool {
	return entry.SWEntry != nil
}

// Is prometheus entry enabled?
func (entry *GinEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// Is common service entry enabled?
func (entry *GinEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// Is TV entry enabled?
func (entry *GinEntry) IsTVEnabled() bool {
	return entry.TVEntry != nil
}

// Is TLS enabled?
func (entry *GinEntry) IsTLSEnabled() bool {
	return entry.CertStore != nil
}

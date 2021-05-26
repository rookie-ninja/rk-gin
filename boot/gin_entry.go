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
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"reflect"
	"runtime"

	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	GinEntryType        = "GinEntry"
	GinEntryDescription = "Internal RK entry which helps to bootstrap with Gin framework."
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
// 6: Gin.TV: See BootConfigTv for details.
// 7: Gin.Prom: See BootConfigProm for details.
// 8: Gin.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 9: Gin.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 10: Gin.Interceptors.BasicAuth.Enabled: Enable basic auth.
// 11: Gin.interceptors.BasicAuth.Credentials: Credential for basic auth, scheme: <user:pass>
// 12: Gin.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 13: Gin.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
type BootConfigGin struct {
	Gin []struct {
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          BootConfigProm          `yaml:"prom" json:"prom"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"metricsProm" json:"metricsProm"`
			BasicAuth struct {
				Enabled     bool     `yaml:"enabled" json:"enabled"`
				Credentials []string `yaml:"credentials" json:"credentials"`
			} `yaml:"basicAuth" json:"basicAuth"`
		} `yaml:"interceptors" json:"interceptors"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"gin" json:"gin"`
}

// GinEntry implements rkentry.Entry interface.
//
// 1: ZapLoggerEntry: See rkentry.ZapLoggerEntry for details.
// 2: EventLoggerEntry: See rkentry.EventLoggerEntry for details.
// 3: Router: gin.Engine created while bootstrapping.
// 4: Server: http.Server created while bootstrapping.
// 5: Port: http/https port server listen to.
// 6: Interceptors: Interceptors user enabled from YAML config, by default, rkginpanic.PanicInterceptor would be injected.
// 7: SwEntry: See SWEntry for details.
// 8: CertEntry: See CertEntry for details..
// 9: CommonServiceEntry: See CommonServiceEntry for details.
// 10: PromEntry: See PromEntry for details.
// 11: TvEntry: See TvEntry for details.
type GinEntry struct {
	EntryName          string                    `json:"entryName" yaml:"entryName"`
	EntryType          string                    `json:"entryType" yaml:"entryType"`
	EntryDescription   string                    `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry   *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	Router             *gin.Engine               `json:"-" yaml:"-"`
	Server             *http.Server              `json:"-" yaml:"-"`
	Port               uint64                    `json:"port" yaml:"port"`
	Interceptors       []gin.HandlerFunc         `json:"-" yaml:"-"`
	SwEntry            *SwEntry                  `json:"swEntry" yaml:"swEntry"`
	CertEntry          *rkentry.CertEntry        `json:"certEntry" yaml:"certEntry"`
	CommonServiceEntry *CommonServiceEntry       `json:"commonServiceEntry" yaml:"commonServiceEntry"`
	PromEntry          *PromEntry                `json:"promEntry" yaml:"promEntry"`
	TvEntry            *TvEntry                  `json:"tvEntry" yaml:"tvEntry"`
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

// Provide TvEntry.
func WithTVEntryGin(tvEntry *TvEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.TvEntry = tvEntry
	}
}

// Provide rkentry.CertEntry.
func WithCertEntryGin(certEntry *rkentry.CertEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CertEntry = certEntry
	}
}

// Provide SwEntry.
func WithSwEntryGin(sw *SwEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.SwEntry = sw
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
		entry.EntryName = name
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

	// 1: Decode config map into boot config struct
	config := &BootConfigGin{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init gin entries with boot config
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

		// Did we enabled swagger?
		var swEntry *SwEntry
		if element.SW.Enabled {
			// Init swagger custom headers from config
			headers := make(map[string]string, 0)
			for i := range element.SW.Headers {
				header := element.SW.Headers[i]
				tokens := strings.Split(header, ":")
				if len(tokens) == 2 {
					headers[tokens[0]] = tokens[1]
				}
			}

			swEntry = NewSwEntry(
				WithNameSw(fmt.Sprintf("%s-sw", element.Name)),
				WithZapLoggerEntrySw(zapLoggerEntry),
				WithEventLoggerEntrySw(eventLoggerEntry),
				WithEnableCommonServiceSw(element.CommonService.Enabled),
				WithPortSw(element.Port),
				WithPathSw(element.SW.Path),
				WithJsonPathSw(element.SW.JsonPath),
				WithHeadersSw(headers))
		}

		// Did we enabled prometheus?
		var promEntry *PromEntry
		if element.Prom.Enabled {
			var pusher *rkprom.PushGatewayPusher
			if element.Prom.Pusher.Enabled {
				certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Prom.Pusher.Cert.Ref)
				var certStore *rkentry.CertStore

				if certEntry != nil {
					certStore = certEntry.Store
				}

				pusher, _ = rkprom.NewPushGatewayPusher(
					rkprom.WithIntervalMSPusher(time.Duration(element.Prom.Pusher.IntervalMs)*time.Millisecond),
					rkprom.WithRemoteAddressPusher(element.Prom.Pusher.RemoteAddress),
					rkprom.WithJobNamePusher(element.Prom.Pusher.JobName),
					rkprom.WithBasicAuthPusher(element.Prom.Pusher.BasicAuth),
					rkprom.WithZapLoggerEntryPusher(zapLoggerEntry),
					rkprom.WithEventLoggerEntryPusher(eventLoggerEntry),
					rkprom.WithCertStorePusher(certStore))
			}

			promEntry = NewPromEntry(
				WithNameProm(fmt.Sprintf("%s-prom", element.Name)),
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithEventLoggerEntryProm(eventLoggerEntry),
				WithPusherProm(pusher))

			if promEntry.Pusher != nil {
				promEntry.Pusher.SetGatherer(promEntry.Gatherer)
			}
		}

		inters := make([]gin.HandlerFunc, 0)

		// Did we enabled logging interceptor?
		if element.Interceptors.LoggingZap.Enabled {
			opts := []rkginlog.Option{
				rkginlog.WithEntryNameAndType(element.Name, "gin"),
				rkginlog.WithEventFactory(eventLoggerEntry.GetEventFactory()),
				rkginlog.WithLogger(zapLoggerEntry.GetLogger()),
			}

			inters = append(inters, rkginlog.LoggingZapInterceptor(opts...))
		}

		// Did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkginmetrics.Option{
				rkginmetrics.WithEntryNameAndType(element.Name, GinEntryType),
			}

			if promEntry != nil {
				opts = append(opts, rkginmetrics.WithRegisterer(promEntry.Registerer))
			}

			inters = append(inters, rkginmetrics.MetricsPromInterceptor(opts...))
		}

		// Did we enabled auth interceptor?
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

		// Did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("%s-commonService", element.Name)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry))
		}

		// Did we enabled tv?
		var tvEntry *TvEntry
		if element.TV.Enabled {
			tvEntry = NewTvEntry(
				WithNameTv(fmt.Sprintf("%s-tv", element.Name)),
				WithZapLoggerEntryTv(zapLoggerEntry),
				WithEventLoggerEntryTv(eventLoggerEntry))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterGinEntry(
			WithZapLoggerEntryGin(zapLoggerEntry),
			WithEventLoggerEntryGin(eventLoggerEntry),
			WithNameGin(name),
			WithPortGin(element.Port),
			WithSwEntryGin(swEntry),
			WithPromEntryGin(promEntry),
			WithCommonServiceEntryGin(commonServiceEntry),
			WithCertEntryGin(certEntry),
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
		EntryType:        GinEntryType,
		EntryDescription: GinEntryDescription,
		Interceptors:     make([]gin.HandlerFunc, 0),
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	entry.Interceptors = append([]gin.HandlerFunc{
		rkginbasic.BasicInterceptor(rkginbasic.WithEntryNameAndType(entry.EntryName, entry.EntryType)),
	}, entry.Interceptors...)

	// insert panic interceptor
	entry.Interceptors = append(entry.Interceptors, rkginpanic.PanicInterceptor())

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "GinServer-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Router == nil {
		gin.SetMode(gin.ReleaseMode)
		entry.Router = gin.New()
	}

	// Default interceptor should be at front
	entry.Router.Use(entry.Interceptors...)

	// Init server only if port is not zero
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
	return entry.EntryName
}

// Get entry type.
func (entry *GinEntry) GetType() string {
	return entry.EntryType
}

// Stringfy gin entry.
func (entry *GinEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// Add basic fields into event.
func (entry *GinEntry) logBasicInfo(event rkquery.Event) {
	event.AddFields(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
		zap.Uint64("port", entry.Port),
		zap.Int("interceptorsCount", len(entry.Interceptors)),
		zap.Bool("swEnabled", entry.IsSwEnabled()),
		zap.Bool("tlsEnabled", entry.IsTlsEnabled()),
		zap.Bool("commonServiceEnabled", entry.IsCommonServiceEnabled()),
		zap.Bool("tvEnabled", entry.IsTvEnabled()),
	)

	if entry.IsSwEnabled() {
		event.AddFields(zap.String("swPath", entry.SwEntry.Path))
	}

	if entry.IsPromEnabled() {
		event.AddFields(
			zap.String("promPath", entry.PromEntry.Path),
			zap.Uint64("promPort", entry.PromEntry.Port))
	}

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
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	ctx = context.Background()

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Router.GET(path.Join(entry.SwEntry.Path, "*any"), entry.SwEntry.ConfigFileHandler())
		entry.Router.GET("/rk/v1/assets/sw/*any", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.Router.GET(entry.PromEntry.Path, gin.WrapH(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))

		// don't start with http handler, we will handle it by ourselves
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Router.GET("/  /healthy", entry.CommonServiceEntry.Healthy)
		entry.Router.GET("/rk/v1/gc", entry.CommonServiceEntry.Gc)
		entry.Router.GET("/rk/v1/info", entry.CommonServiceEntry.Info)
		entry.Router.GET("/rk/v1/configs", entry.CommonServiceEntry.Configs)
		entry.Router.GET("/rk/v1/apis", entry.CommonServiceEntry.Apis)
		entry.Router.GET("/rk/v1/sys", entry.CommonServiceEntry.Sys)
		entry.Router.GET("/rk/v1/req", entry.CommonServiceEntry.Req)
		entry.Router.GET("/rk/v1/entries", entry.CommonServiceEntry.Entries)
		entry.Router.GET("/rk/v1/certs", entry.CommonServiceEntry.Certs)
		entry.Router.GET("/rk/v1/logs", entry.CommonServiceEntry.Logs)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Router.RouterGroup.GET("/rk/v1/tv/*item", entry.TvEntry.TV)
		entry.Router.GET("/rk/v1/assets/tv/*any", entry.TvEntry.AssetsFileHandler())

		entry.TvEntry.Bootstrap(ctx)
	}

	// Start gin server
	if entry.Server != nil {
		entry.ZapLoggerEntry.GetLogger().Info("Bootstrapping GinEntry.", event.GetFields()...)
		entry.EventLoggerEntry.GetEventHelper().Finish(event)

		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			if cert, err := tls.X509KeyPair(entry.CertEntry.Store.ServerCert, entry.CertEntry.Store.ServerKey); err != nil {
				event.AddErr(err)
				entry.ZapLoggerEntry.GetLogger().Error("Error occurs while parsing TLS.", event.GetFields()...)
				entry.EventLoggerEntry.GetEventHelper().Finish(event)
				rkcommon.ShutdownWithError(err)
			} else {
				entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
			}

			if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				entry.ZapLoggerEntry.GetLogger().Error("Error occurs while serving gin-listener-tls.", event.GetFields()...)
				entry.EventLoggerEntry.GetEventHelper().Finish(event)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				entry.ZapLoggerEntry.GetLogger().Error("Error occurs while serving gin-listener.", event.GetFields()...)
				entry.EventLoggerEntry.GetEventHelper().Finish(event)
				rkcommon.ShutdownWithError(err)
			}
		}
	}
}

func (entry *GinEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

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
		entry.ZapLoggerEntry.GetLogger().Info("Interrupting GinEntry.", event.GetFields()...)
		if err := entry.Server.Shutdown(context.Background()); err != nil {
			event.AddErr(err)
			entry.ZapLoggerEntry.GetLogger().Warn("Error occurs while stopping gin-server.", event.GetFields()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Is swagger entry enabled?
func (entry *GinEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
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
func (entry *GinEntry) IsTvEnabled() bool {
	return entry.TvEntry != nil
}

// Is TLS enabled?
func (entry *GinEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// Marshal entry.
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

	interceptorsStr := make([]string, 0)
	m["interceptors"] = &interceptorsStr

	for i := range entry.Interceptors {
		element := entry.Interceptors[i]
		interceptorsStr = append(interceptorsStr,
			path.Base(runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name()))
	}

	return json.Marshal(&m)
}

// Not supported.
func (entry *GinEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Unmarshal entry.
func (entry *GinEntry) GetDescription() string {
	return entry.EntryDescription
}

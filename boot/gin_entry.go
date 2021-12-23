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
	"github.com/markbates/pkger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/cors"
	"github.com/rookie-ninja/rk-gin/interceptor/csrf"
	"github.com/rookie-ninja/rk-gin/interceptor/gzip"
	"github.com/rookie-ninja/rk-gin/interceptor/jwt"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/meta"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-gin/interceptor/panic"
	"github.com/rookie-ninja/rk-gin/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-gin/interceptor/secure"
	"github.com/rookie-ninja/rk-gin/interceptor/timeout"
	"github.com/rookie-ninja/rk-gin/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"os"
	"path/filepath"
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
	// GinEntryType type of entry
	GinEntryType = "GinEntry"
	// GinEntryDescription description of entry
	GinEntryDescription = "Internal RK entry which helps to bootstrap with Gin framework."
)

var bootstrapEventIdKey = eventIdKey{}

type eventIdKey struct{}

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap gin entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterGinEntriesWithConfig)
}

// BootConfigGin boot config which is for gin entry.
//
// 1: Gin.Enabled: Enable gin entry, default is true.
// 2: Gin.Name: Name of gin entry, should be unique globally.
// 3: Gin.Port: Port of gin entry.
// 4: Gin.Cert.Ref: Reference of rkentry.CertEntry.
// 5: Gin.SW: See BootConfigSW for details.
// 6: Gin.CommonService: See BootConfigCommonService for details.
// 7: Gin.TV: See BootConfigTv for details.
// 8: Gin.Prom: See BootConfigProm for details.
// 9: Gin.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 10: Gin.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 11: Gin.Interceptors.auth.Enabled: Enable basic auth.
// 12: Gin.Interceptors.auth.Basic: Credential for basic auth, scheme: <user:pass>
// 13: Gin.Interceptors.auth.ApiKey: Credential for X-API-Key.
// 14: Gin.Interceptors.auth.igorePrefix: List of paths that will be ignored.
// 15: Gin.Interceptors.Extension.Enabled: Enable extension interceptor.
// 16: Gin.Interceptors.Extension.Prefix: Prefix of extension header key.
// 17: Gin.Interceptors.TracingTelemetry.Enabled: Enable tracing interceptor with opentelemetry.
// 18: Gin.Interceptors.TracingTelemetry.Exporter.File.Enabled: Enable file exporter which support type of stdout and local file.
// 19: Gin.Interceptors.TracingTelemetry.Exporter.File.OutputPath: Output path of file exporter, stdout and file path is supported.
// 20: Gin.Interceptors.TracingTelemetry.Exporter.Jaeger.Enabled: Enable jaeger exporter.
// 21: Gin.Interceptors.TracingTelemetry.Exporter.Jaeger.AgentEndpoint: Specify jeager agent endpoint, localhost:6832 would be used by default.
// 22: Gin.Interceptors.RateLimit.Enabled: Enable rate limit interceptor.
// 23: Gin.Interceptors.RateLimit.Algorithm: Algorithm of rate limiter.
// 24: Gin.Interceptors.RateLimit.ReqPerSec: Request per second.
// 25: Gin.Interceptors.RateLimit.Paths.path: Name of full path.
// 26: Gin.Interceptors.RateLimit.Paths.ReqPerSec: Request per second by path.
// 27: Gin.Interceptors.Timeout.Enabled: Enable timeout interceptor.
// 28: Gin.Interceptors.Timeout.TimeoutMs: Timeout in milliseconds.
// 29: Gin.Interceptors.Timeout.Paths.path: Name of full path.
// 30: Gin.Interceptors.Timeout.Paths.TimeoutMs: Timeout in milliseconds by path.
// 31: Gin.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 32: Gin.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
type BootConfigGin struct {
	Gin []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
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
		Static        BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled                bool     `yaml:"enabled" json:"enabled"`
				ZapLoggerEncoding      string   `yaml:"zapLoggerEncoding" json:"zapLoggerEncoding"`
				ZapLoggerOutputPaths   []string `yaml:"zapLoggerOutputPaths" json:"zapLoggerOutputPaths"`
				EventLoggerEncoding    string   `yaml:"eventLoggerEncoding" json:"eventLoggerEncoding"`
				EventLoggerOutputPaths []string `yaml:"eventLoggerOutputPaths" json:"eventLoggerOutputPaths"`
			} `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"metricsProm" json:"metricsProm"`
			Auth struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				Basic        []string `yaml:"basic" json:"basic"`
				ApiKey       []string `yaml:"apiKey" json:"apiKey"`
			} `yaml:"auth" json:"auth"`
			Cors struct {
				Enabled          bool     `yaml:"enabled" json:"enabled"`
				AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
				AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
				AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
				AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
				ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
				MaxAge           int      `yaml:"maxAge" json:"maxAge"`
			} `yaml:"cors" json:"cors"`
			Meta struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Prefix  string `yaml:"prefix" json:"prefix"`
			} `yaml:"meta" json:"meta"`
			Jwt struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				SigningKey   string   `yaml:"signingKey" json:"signingKey"`
				SigningKeys  []string `yaml:"signingKeys" json:"signingKeys"`
				SigningAlgo  string   `yaml:"signingAlgo" json:"signingAlgo"`
				TokenLookup  string   `yaml:"tokenLookup" json:"tokenLookup"`
				AuthScheme   string   `yaml:"authScheme" json:"authScheme"`
			} `yaml:"jwt" json:"jwt"`
			Secure struct {
				Enabled               bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix          []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				XssProtection         string   `yaml:"xssProtection" json:"xssProtection"`
				ContentTypeNosniff    string   `yaml:"contentTypeNosniff" json:"contentTypeNosniff"`
				XFrameOptions         string   `yaml:"xFrameOptions" json:"xFrameOptions"`
				HstsMaxAge            int      `yaml:"hstsMaxAge" json:"hstsMaxAge"`
				HstsExcludeSubdomains bool     `yaml:"hstsExcludeSubdomains" json:"hstsExcludeSubdomains"`
				HstsPreloadEnabled    bool     `yaml:"hstsPreloadEnabled" json:"hstsPreloadEnabled"`
				ContentSecurityPolicy string   `yaml:"contentSecurityPolicy" json:"contentSecurityPolicy"`
				CspReportOnly         bool     `yaml:"cspReportOnly" json:"cspReportOnly"`
				ReferrerPolicy        string   `yaml:"referrerPolicy" json:"referrerPolicy"`
			} `yaml:"secure" json:"secure"`
			RateLimit struct {
				Enabled   bool   `yaml:"enabled" json:"enabled"`
				Algorithm string `yaml:"algorithm" json:"algorithm"`
				ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"rateLimit" json:"rateLimit"`
			Csrf struct {
				Enabled        bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix   []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				TokenLength    int      `yaml:"tokenLength" json:"tokenLength"`
				TokenLookup    string   `yaml:"tokenLookup" json:"tokenLookup"`
				CookieName     string   `yaml:"cookieName" json:"cookieName"`
				CookieDomain   string   `yaml:"cookieDomain" json:"cookieDomain"`
				CookiePath     string   `yaml:"cookiePath" json:"cookiePath"`
				CookieMaxAge   int      `yaml:"cookieMaxAge" json:"cookieMaxAge"`
				CookieHttpOnly bool     `yaml:"cookieHttpOnly" json:"cookieHttpOnly"`
				CookieSameSite string   `yaml:"cookieSameSite" json:"cookieSameSite"`
			} `yaml:"csrf" yaml:"csrf"`
			Gzip struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Level   string `yaml:"level" json:"level"`
			} `yaml:"gzip" json:"gzip"`
			Timeout struct {
				Enabled   bool `yaml:"enabled" json:"enabled"`
				TimeoutMs int  `yaml:"timeoutMs" json:"timeoutMs"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					TimeoutMs int    `yaml:"timeoutMs" json:"timeoutMs"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"timeout" json:"timeout"`
			TracingTelemetry struct {
				Enabled  bool `yaml:"enabled" json:"enabled"`
				Exporter struct {
					File struct {
						Enabled    bool   `yaml:"enabled" json:"enabled"`
						OutputPath string `yaml:"outputPath" json:"outputPath"`
					} `yaml:"file" json:"file"`
					Jaeger struct {
						Agent struct {
							Enabled bool   `yaml:"enabled" json:"enabled"`
							Host    string `yaml:"host" json:"host"`
							Port    int    `yaml:"port" json:"port"`
						} `yaml:"agent" json:"agent"`
						Collector struct {
							Enabled  bool   `yaml:"enabled" json:"enabled"`
							Endpoint string `yaml:"endpoint" json:"endpoint"`
							Username string `yaml:"username" json:"username"`
							Password string `yaml:"password" json:"password"`
						} `yaml:"collector" json:"collector"`
					} `yaml:"jaeger" json:"jaeger"`
				} `yaml:"exporter" json:"exporter"`
			} `yaml:"tracingTelemetry" json:"tracingTelemetry"`
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
	StaticFileEntry    *StaticFileHandlerEntry   `json:"staticFileHandlerEntry" yaml:"staticFileHandlerEntry"`
	TvEntry            *TvEntry                  `json:"tvEntry" yaml:"tvEntry"`
}

// GinEntryOption Gin entry option.
type GinEntryOption func(*GinEntry)

// GetGinEntry Get GinEntry from rkentry.GlobalAppCtx.
func GetGinEntry(name string) *GinEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*GinEntry)
	return entry
}

// WithZapLoggerEntryGin provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryGin(zapLogger *rkentry.ZapLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntryGin provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryGin(eventLogger *rkentry.EventLoggerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithInterceptorsGin provide user interceptors.
func WithInterceptorsGin(inters ...gin.HandlerFunc) GinEntryOption {
	return func(entry *GinEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]gin.HandlerFunc, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// WithCommonServiceEntryGin provide CommonServiceEntry.
func WithCommonServiceEntryGin(commonServiceEntry *CommonServiceEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithTVEntryGin provide TvEntry.
func WithTVEntryGin(tvEntry *TvEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.TvEntry = tvEntry
	}
}

// WithStaticFileHandlerEntryGin provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntryGin(staticEntry *StaticFileHandlerEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithCertEntryGin provide rkentry.CertEntry.
func WithCertEntryGin(certEntry *rkentry.CertEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntryGin provide SwEntry.
func WithSwEntryGin(sw *SwEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.SwEntry = sw
	}
}

// WithPortGin provide port.
func WithPortGin(port uint64) GinEntryOption {
	return func(entry *GinEntry) {
		entry.Port = port
	}
}

// WithNameGin provide name.
func WithNameGin(name string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionGin provide name.
func WithDescriptionGin(description string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.EntryDescription = description
	}
}

// WithPromEntryGin provide PromEntry.
func WithPromEntryGin(prom *PromEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.PromEntry = prom
	}
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
	config := &BootConfigGin{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init gin entries with boot config
	for i := range config.Gin {
		element := config.Gin[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		promRegistry := prometheus.NewRegistry()
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

			promRegistry.Register(prometheus.NewGoCollector())
			promEntry = NewPromEntry(
				WithNameProm(fmt.Sprintf("%s-prom", element.Name)),
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithPromRegistryProm(promRegistry),
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
				rkginlog.WithEntryNameAndType(element.Name, GinEntryType),
				rkginlog.WithEventLoggerEntry(eventLoggerEntry),
				rkginlog.WithZapLoggerEntry(zapLoggerEntry),
			}

			if strings.ToLower(element.Interceptors.LoggingZap.ZapLoggerEncoding) == "json" {
				opts = append(opts, rkginlog.WithZapLoggerEncoding(rkginlog.ENCODING_JSON))
			}

			if strings.ToLower(element.Interceptors.LoggingZap.EventLoggerEncoding) == "json" {
				opts = append(opts, rkginlog.WithEventLoggerEncoding(rkginlog.ENCODING_JSON))
			}

			if len(element.Interceptors.LoggingZap.ZapLoggerOutputPaths) > 0 {
				opts = append(opts, rkginlog.WithZapLoggerOutputPaths(element.Interceptors.LoggingZap.ZapLoggerOutputPaths...))
			}

			if len(element.Interceptors.LoggingZap.EventLoggerOutputPaths) > 0 {
				opts = append(opts, rkginlog.WithEventLoggerOutputPaths(element.Interceptors.LoggingZap.EventLoggerOutputPaths...))
			}

			inters = append(inters, rkginlog.Interceptor(opts...))
		}

		// Did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkginmetrics.Option{
				rkginmetrics.WithRegisterer(promRegistry),
				rkginmetrics.WithEntryNameAndType(element.Name, GinEntryType),
			}

			inters = append(inters, rkginmetrics.Interceptor(opts...))
		}

		// Did we enabled tracing interceptor?
		if element.Interceptors.TracingTelemetry.Enabled {
			var exporter trace.SpanExporter

			if element.Interceptors.TracingTelemetry.Exporter.File.Enabled {
				exporter = rkgintrace.CreateFileExporter(element.Interceptors.TracingTelemetry.Exporter.File.OutputPath)
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Enabled {
				opts := make([]jaeger.AgentEndpointOption, 0)
				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host) > 0 {
					opts = append(opts,
						jaeger.WithAgentHost(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host))
				}
				if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port > 0 {
					opts = append(opts,
						jaeger.WithAgentPort(
							fmt.Sprintf("%d", element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port)))
				}

				exporter = rkgintrace.CreateJaegerExporter(jaeger.WithAgentEndpoint(opts...))
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Enabled {
				opts := []jaeger.CollectorEndpointOption{
					jaeger.WithUsername(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Username),
					jaeger.WithPassword(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Password),
				}

				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint) > 0 {
					opts = append(opts, jaeger.WithEndpoint(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint))
				}

				exporter = rkgintrace.CreateJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
			}

			opts := []rkgintrace.Option{
				rkgintrace.WithEntryNameAndType(element.Name, GinEntryType),
				rkgintrace.WithExporter(exporter),
			}

			inters = append(inters, rkgintrace.Interceptor(opts...))
		}

		// Did we enabled jwt interceptor?
		if element.Interceptors.Jwt.Enabled {
			var signingKey []byte
			if len(element.Interceptors.Jwt.SigningKey) > 0 {
				signingKey = []byte(element.Interceptors.Jwt.SigningKey)
			}

			opts := []rkginjwt.Option{
				rkginjwt.WithEntryNameAndType(element.Name, GinEntryType),
				rkginjwt.WithSigningKey(signingKey),
				rkginjwt.WithSigningAlgorithm(element.Interceptors.Jwt.SigningAlgo),
				rkginjwt.WithTokenLookup(element.Interceptors.Jwt.TokenLookup),
				rkginjwt.WithAuthScheme(element.Interceptors.Jwt.AuthScheme),
				rkginjwt.WithIgnorePrefix(element.Interceptors.Jwt.IgnorePrefix...),
			}

			for _, v := range element.Interceptors.Jwt.SigningKeys {
				tokens := strings.SplitN(v, ":", 2)
				if len(tokens) == 2 {
					opts = append(opts, rkginjwt.WithSigningKeys(tokens[0], tokens[1]))
				}
			}

			inters = append(inters, rkginjwt.Interceptor(opts...))
		}

		// Did we enabled secure interceptor?
		if element.Interceptors.Secure.Enabled {
			opts := []rkginsec.Option{
				rkginsec.WithEntryNameAndType(element.Name, GinEntryType),
				rkginsec.WithXSSProtection(element.Interceptors.Secure.XssProtection),
				rkginsec.WithContentTypeNosniff(element.Interceptors.Secure.ContentTypeNosniff),
				rkginsec.WithXFrameOptions(element.Interceptors.Secure.XFrameOptions),
				rkginsec.WithHSTSMaxAge(element.Interceptors.Secure.HstsMaxAge),
				rkginsec.WithHSTSExcludeSubdomains(element.Interceptors.Secure.HstsExcludeSubdomains),
				rkginsec.WithHSTSPreloadEnabled(element.Interceptors.Secure.HstsPreloadEnabled),
				rkginsec.WithContentSecurityPolicy(element.Interceptors.Secure.ContentSecurityPolicy),
				rkginsec.WithCSPReportOnly(element.Interceptors.Secure.CspReportOnly),
				rkginsec.WithReferrerPolicy(element.Interceptors.Secure.ReferrerPolicy),
				rkginsec.WithIgnorePrefix(element.Interceptors.Secure.IgnorePrefix...),
			}

			inters = append(inters, rkginsec.Interceptor(opts...))
		}

		// Did we enabled csrf interceptor?
		if element.Interceptors.Csrf.Enabled {
			opts := []rkgincsrf.Option{
				rkgincsrf.WithEntryNameAndType(element.Name, GinEntryType),
				rkgincsrf.WithTokenLength(element.Interceptors.Csrf.TokenLength),
				rkgincsrf.WithTokenLookup(element.Interceptors.Csrf.TokenLookup),
				rkgincsrf.WithCookieName(element.Interceptors.Csrf.CookieName),
				rkgincsrf.WithCookieDomain(element.Interceptors.Csrf.CookieDomain),
				rkgincsrf.WithCookiePath(element.Interceptors.Csrf.CookiePath),
				rkgincsrf.WithCookieMaxAge(element.Interceptors.Csrf.CookieMaxAge),
				rkgincsrf.WithCookieHTTPOnly(element.Interceptors.Csrf.CookieHttpOnly),
				rkgincsrf.WithIgnorePrefix(element.Interceptors.Csrf.IgnorePrefix...),
			}

			// convert to string to cookie same sites
			sameSite := http.SameSiteDefaultMode

			switch strings.ToLower(element.Interceptors.Csrf.CookieSameSite) {
			case "lax":
				sameSite = http.SameSiteLaxMode
			case "strict":
				sameSite = http.SameSiteStrictMode
			case "none":
				sameSite = http.SameSiteNoneMode
			default:
				sameSite = http.SameSiteDefaultMode
			}

			opts = append(opts, rkgincsrf.WithCookieSameSite(sameSite))

			inters = append(inters, rkgincsrf.Interceptor(opts...))
		}

		// Did we enabled cors interceptor?
		if element.Interceptors.Cors.Enabled {
			opts := []rkgincors.Option{
				rkgincors.WithEntryNameAndType(element.Name, GinEntryType),
				rkgincors.WithAllowOrigins(element.Interceptors.Cors.AllowOrigins...),
				rkgincors.WithAllowCredentials(element.Interceptors.Cors.AllowCredentials),
				rkgincors.WithExposeHeaders(element.Interceptors.Cors.ExposeHeaders...),
				rkgincors.WithMaxAge(element.Interceptors.Cors.MaxAge),
				rkgincors.WithAllowHeaders(element.Interceptors.Cors.AllowHeaders...),
				rkgincors.WithAllowMethods(element.Interceptors.Cors.AllowMethods...),
			}

			inters = append(inters, rkgincors.Interceptor(opts...))
		}

		// Did we enabled gzip interceptor?
		if element.Interceptors.Gzip.Enabled {
			opts := []rkgingzip.Option{
				rkgingzip.WithEntryNameAndType(element.Name, GinEntryType),
				rkgingzip.WithLevel(element.Interceptors.Gzip.Level),
			}

			inters = append(inters, rkgingzip.Interceptor(opts...))
		}

		// Did we enabled meta interceptor?
		if element.Interceptors.Meta.Enabled {
			opts := []rkginmeta.Option{
				rkginmeta.WithEntryNameAndType(element.Name, GinEntryType),
				rkginmeta.WithPrefix(element.Interceptors.Meta.Prefix),
			}

			inters = append(inters, rkginmeta.Interceptor(opts...))
		}

		// Did we enabled auth interceptor?
		if element.Interceptors.Auth.Enabled {
			opts := make([]rkginauth.Option, 0)
			opts = append(opts,
				rkginauth.WithEntryNameAndType(element.Name, GinEntryType),
				rkginauth.WithBasicAuth(element.Name, element.Interceptors.Auth.Basic...),
				rkginauth.WithApiKeyAuth(element.Interceptors.Auth.ApiKey...))

			// Add exceptional path
			if swEntry != nil {
				opts = append(opts, rkginauth.WithIgnorePrefix(strings.TrimSuffix(swEntry.Path, "/")))
			}

			opts = append(opts, rkginauth.WithIgnorePrefix("/rk/v1/assets"))
			opts = append(opts, rkginauth.WithIgnorePrefix(element.Interceptors.Auth.IgnorePrefix...))

			inters = append(inters, rkginauth.Interceptor(opts...))
		}

		// Did we enabled timeout interceptor?
		// This should be in front of rate limit interceptor since rate limit may block over the threshold of timeout.
		if element.Interceptors.Timeout.Enabled {
			opts := make([]rkgintimeout.Option, 0)
			opts = append(opts,
				rkgintimeout.WithEntryNameAndType(element.Name, GinEntryType))

			timeout := time.Duration(element.Interceptors.Timeout.TimeoutMs) * time.Millisecond
			opts = append(opts, rkgintimeout.WithTimeoutAndResp(timeout, nil))

			for i := range element.Interceptors.Timeout.Paths {
				e := element.Interceptors.Timeout.Paths[i]
				timeout := time.Duration(e.TimeoutMs) * time.Millisecond
				opts = append(opts, rkgintimeout.WithTimeoutAndRespByPath(e.Path, timeout, nil))
			}

			inters = append(inters, rkgintimeout.Interceptor(opts...))
		}

		// Did we enabled rate limit interceptor?
		if element.Interceptors.RateLimit.Enabled {
			opts := make([]rkginlimit.Option, 0)
			opts = append(opts,
				rkginlimit.WithEntryNameAndType(element.Name, GinEntryType))

			if len(element.Interceptors.RateLimit.Algorithm) > 0 {
				opts = append(opts, rkginlimit.WithAlgorithm(element.Interceptors.RateLimit.Algorithm))
			}
			opts = append(opts, rkginlimit.WithReqPerSec(element.Interceptors.RateLimit.ReqPerSec))

			for i := range element.Interceptors.RateLimit.Paths {
				e := element.Interceptors.RateLimit.Paths[i]
				opts = append(opts, rkginlimit.WithReqPerSecByPath(e.Path, e.ReqPerSec))
			}

			inters = append(inters, rkginlimit.Interceptor(opts...))
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

		// DId we enabled static file handler?
		var staticEntry *StaticFileHandlerEntry
		if element.Static.Enabled {
			var fs http.FileSystem
			switch element.Static.SourceType {
			case "pkger":
				fs = pkger.Dir(element.Static.SourcePath)
			case "local":
				if !filepath.IsAbs(element.Static.SourcePath) {
					wd, _ := os.Getwd()
					element.Static.SourcePath = path.Join(wd, element.Static.SourcePath)
				}
				fs = http.Dir(element.Static.SourcePath)
			}

			staticEntry = NewStaticFileHandlerEntry(
				WithZapLoggerEntryStatic(zapLoggerEntry),
				WithEventLoggerEntryStatic(eventLoggerEntry),
				WithNameStatic(fmt.Sprintf("%s-static", element.Name)),
				WithPathStatic(element.Static.Path),
				WithFileSystemStatic(fs))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterGinEntry(
			WithZapLoggerEntryGin(zapLoggerEntry),
			WithEventLoggerEntryGin(eventLoggerEntry),
			WithNameGin(name),
			WithDescriptionGin(element.Description),
			WithPortGin(element.Port),
			WithSwEntryGin(swEntry),
			WithPromEntryGin(promEntry),
			WithCommonServiceEntryGin(commonServiceEntry),
			WithCertEntryGin(certEntry),
			WithTVEntryGin(tvEntry),
			WithStaticFileHandlerEntryGin(staticEntry),
			WithInterceptorsGin(inters...))

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
		Interceptors:     make([]gin.HandlerFunc, 0),
		Port:             80,
	}

	for i := range opts {
		opts[i](entry)
	}

	// insert panic interceptor
	entry.Interceptors = append(entry.Interceptors, rkginpanic.Interceptor(
		rkginpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

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

	if entry.Port != 0 {
		entry.Server = &http.Server{
			Addr:    "0.0.0.0:" + strconv.FormatUint(entry.Port, 10),
			Handler: entry.Router,
		}
	}

	// Default interceptor should be at front
	entry.Router.Use(entry.Interceptors...)

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

// String Stringfy gin entry.
func (entry *GinEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// Add basic fields into event.
func (entry *GinEntry) logBasicInfo(event rkquery.Event) {
	event.AddPayloads(
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
		event.AddPayloads()
		event.AddPayloads(zap.String("swPath", entry.SwEntry.Path))
	}

	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.String("promPath", entry.PromEntry.Path),
			zap.Uint64("promPort", entry.PromEntry.Port))
	}

}

// Bootstrap GinEntry.
func (entry *GinEntry) Bootstrap(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Router.GET(path.Join(entry.SwEntry.Path, "*any"), entry.SwEntry.ConfigFileHandler())
		entry.Router.GET("/rk/v1/assets/sw/*any", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		// Register path into Router.
		entry.Router.GET(path.Join(entry.StaticFileEntry.Path, "*any"), entry.StaticFileEntry.GetFileHandler())

		// Bootstrap entry.
		entry.StaticFileEntry.Bootstrap(ctx)
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
		entry.Router.GET("/rk/v1/healthy", entry.CommonServiceEntry.Healthy)
		entry.Router.GET("/rk/v1/gc", entry.CommonServiceEntry.Gc)
		entry.Router.GET("/rk/v1/info", entry.CommonServiceEntry.Info)
		entry.Router.GET("/rk/v1/configs", entry.CommonServiceEntry.Configs)
		entry.Router.GET("/rk/v1/apis", entry.CommonServiceEntry.Apis)
		entry.Router.GET("/rk/v1/sys", entry.CommonServiceEntry.Sys)
		entry.Router.GET("/rk/v1/req", entry.CommonServiceEntry.Req)
		entry.Router.GET("/rk/v1/entries", entry.CommonServiceEntry.Entries)
		entry.Router.GET("/rk/v1/certs", entry.CommonServiceEntry.Certs)
		entry.Router.GET("/rk/v1/logs", entry.CommonServiceEntry.Logs)
		entry.Router.GET("/rk/v1/deps", entry.CommonServiceEntry.Deps)
		entry.Router.GET("/rk/v1/license", entry.CommonServiceEntry.License)
		entry.Router.GET("/rk/v1/readme", entry.CommonServiceEntry.Readme)
		entry.Router.GET("/rk/v1/git", entry.CommonServiceEntry.Git)

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
	logger.Info("Bootstrapping GinEntry.", event.ListPayloads()...)
	go func(*GinEntry) {
		if entry.Server != nil {
			// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
			if entry.IsTlsEnabled() {
				if cert, err := tls.X509KeyPair(entry.CertEntry.Store.ServerCert, entry.CertEntry.Store.ServerKey); err != nil {
					event.AddErr(err)
					logger.Error("Error occurs while parsing TLS.", event.ListPayloads()...)
					rkcommon.ShutdownWithError(err)
				} else {
					entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
				}

				if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
					event.AddErr(err)
					logger.Error("Error occurs while serving gin-listener-tls.", event.ListPayloads()...)
					rkcommon.ShutdownWithError(err)
				}
			} else {
				if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					event.AddErr(err)
					logger.Error("Error occurs while serving gin-listener.", event.ListPayloads()...)
					rkcommon.ShutdownWithError(err)
				}
			}
		}
	}(entry)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt GinEntry.
func (entry *GinEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	entry.logBasicInfo(event)

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
		logger.Info("Interrupting GinEntry.", event.ListPayloads()...)
		if err := entry.Server.Shutdown(context.Background()); err != nil {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping gin-server.", event.ListPayloads()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *GinEntry) AddInterceptor(inters ...gin.HandlerFunc) {
	entry.Interceptors = append(entry.Interceptors, inters...)
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

	interceptorsStr := make([]string, 0)
	m["interceptors"] = &interceptorsStr

	for i := range entry.Interceptors {
		element := entry.Interceptors[i]
		interceptorsStr = append(interceptorsStr,
			path.Base(runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name()))
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *GinEntry) UnmarshalJSON([]byte) error {
	return nil
}

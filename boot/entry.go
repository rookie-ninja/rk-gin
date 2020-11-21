// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/context"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type bootConfig struct {
	Gin []struct {
		Name string `yaml:"name"`
		Port uint64 `yaml:"port"`
		TLS  struct {
			Enabled bool `yaml:"enabled"`
			User    struct {
				Enabled  bool   `yaml:"enabled"`
				CertFile string `yaml:"certFile"`
				KeyFile  string `yaml:"keyFile"`
			} `yaml:"user"`
			Auto struct {
				Enabled    bool   `yaml:"enabled"`
				CertOutput string `yaml:"certOutput"`
			} `yaml:"auto"`
		} `yaml:"tls"`
		SW struct {
			Enabled  bool     `yaml:"enabled"`
			Path     string   `yaml:"path"`
			JSONPath string   `yaml:"jsonPath"`
			Headers  []string `yaml:"headers"`
		} `yaml:"sw"`
		EnableCommonService bool `yaml:"enableCommonService"`
		EnableTV            bool `yaml:"enableTB"`
		LoggingInterceptor  struct {
			Enabled       bool `yaml:"enabled"`
			EnableLogging bool `yaml:"enableLogging"`
			EnableMetrics bool `yaml:"enableMetrics"`
		} `yaml:"loggingInterceptor"`
		AuthInterceptor struct {
			Enabled     bool     `yaml:"enabled"`
			Realm       string   `yaml:"realm"`
			Credentials []string `yaml:"credentials"`
		} `yaml:"authInterceptor"`
	} `yaml:"gin"`
}

type GinEntry struct {
	logger              *zap.Logger
	eventFactory        *rk_query.EventFactory
	router              *gin.Engine
	server              *http.Server
	name                string
	port                uint64
	interceptors        []gin.HandlerFunc
	sw                  *swEntry
	tls                 *tlsEntry
	enableCommonService bool
	enableTV            bool
	entryType           string
}

type GinEntryOption func(*GinEntry)

func WithLogger(logger *zap.Logger) GinEntryOption {
	return func(entry *GinEntry) {
		entry.logger = logger
	}
}

func WithEventFactory(factory *rk_query.EventFactory) GinEntryOption {
	return func(entry *GinEntry) {
		entry.eventFactory = factory
	}
}

func WithInterceptors(inters ...gin.HandlerFunc) GinEntryOption {
	return func(entry *GinEntry) {
		if entry.interceptors == nil {
			entry.interceptors = make([]gin.HandlerFunc, 0)
		}

		entry.interceptors = append(entry.interceptors, inters...)
	}
}

func WithEnableCommonService(enable bool) GinEntryOption {
	return func(entry *GinEntry) {
		entry.enableCommonService = enable
	}
}

func WithEnableTV(enable bool) GinEntryOption {
	return func(entry *GinEntry) {
		entry.enableTV = enable
	}
}

func WithRouter(router *gin.Engine) GinEntryOption {
	return func(entry *GinEntry) {
		entry.router = router
	}
}

func WithTlsEntry(tls *tlsEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.tls = tls
	}
}

func WithSWEntry(sw *swEntry) GinEntryOption {
	return func(entry *GinEntry) {
		entry.sw = sw
	}
}

func WithPort(port uint64) GinEntryOption {
	return func(entry *GinEntry) {
		entry.port = port
	}
}

func WithName(name string) GinEntryOption {
	return func(entry *GinEntry) {
		entry.name = name
	}
}

func NewGinEntries(path string, factory *rk_query.EventFactory, logger *zap.Logger) map[string]*GinEntry {
	bytes := readFile(path)
	config := &bootConfig{}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil
	}

	return getGinServerEntries(config, factory, logger)
}

func getGinServerEntries(config *bootConfig, factory *rk_query.EventFactory, logger *zap.Logger) map[string]*GinEntry {
	res := make(map[string]*GinEntry)

	for i := range config.Gin {
		element := config.Gin[i]
		name := element.Name

		// did we enabled swagger?
		var swEntry *swEntry
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

			swEntry = newSWEntry(
				withPort(element.Port),
				withPath(element.SW.Path),
				withJsonPath(element.SW.JSONPath),
				withHeaders(headers))
		}

		// did we enabled tls?
		var tlsEntry *tlsEntry
		if element.TLS.Enabled {
			if element.TLS.User.Enabled {
				tlsEntry = newTlsEntry(
					withCertFilePath(element.TLS.User.CertFile),
					withKeyFilePath(element.TLS.User.KeyFile))
			} else if element.TLS.Auto.Enabled {
				tlsEntry = newTlsEntry(
					withGenerateCert(element.TLS.Auto.Enabled),
					withGeneratePath(element.TLS.Auto.CertOutput))
			}
		}

		inters := make([]gin.HandlerFunc, 0)
		// did we enabled logging interceptor?
		if element.LoggingInterceptor.Enabled {
			opts := []rk_gin_log.Option{
				rk_gin_log.WithEventFactory(factory),
				rk_gin_log.WithLogger(logger),
				rk_gin_log.WithEnableLogging(element.LoggingInterceptor.EnableLogging),
				rk_gin_log.WithEnableMetrics(element.LoggingInterceptor.EnableMetrics),
			}

			inters = append(inters, rk_gin_log.RkGinLog(opts...))
		}

		// did we enabled auth interceptor?
		if element.AuthInterceptor.Enabled {
			accounts := gin.Accounts{}
			for i := range element.AuthInterceptor.Credentials {
				cred := element.AuthInterceptor.Credentials[i]
				tokens := strings.Split(cred, ":")
				if len(tokens) == 2 {
					accounts[tokens[0]] = tokens[1]
				}
			}
			inters = append(inters, rk_gin_auth.RkGinAuth(accounts, element.AuthInterceptor.Realm))
		}

		entry := NewGinEntry(
			WithName(name),
			WithPort(element.Port),
			WithSWEntry(swEntry),
			WithTlsEntry(tlsEntry),
			WithInterceptors(inters...),
			WithEnableCommonService(element.EnableCommonService),
			WithEnableTV(element.EnableTV))

		res[name] = entry
	}

	return res
}

func NewGinEntry(opts ...GinEntryOption) *GinEntry {
	entry := &GinEntry{
		logger:       rk_logger.StdoutLogger,
		eventFactory: rk_query.NewEventFactory(),
		entryType:    "gin",
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.logger == nil {
		entry.logger = rk_logger.StdoutLogger
	}

	if entry.eventFactory == nil {
		entry.eventFactory = rk_query.NewEventFactory()
	}

	if len(entry.name) < 1 {
		entry.name = "gin-server-" + strconv.FormatUint(entry.port, 10)
	}

	if entry.interceptors == nil {
		entry.interceptors = make([]gin.HandlerFunc, 0)
	}

	if entry.router == nil {
		gin.SetMode(gin.ReleaseMode)
		entry.router = gin.New()
	}

	if entry.sw != nil {
		entry.router.GET(path.Join(entry.sw.getPath(), "*any"), entry.sw.ginHandler())
		entry.router.GET("/swagger/*any", entry.sw.ginFileHandler())
	}

	entry.interceptors = append(entry.interceptors, rk_gin_panic.RkGinPanic())
	entry.router.Use(entry.interceptors...)

	if entry.enableCommonService {
		entry.GetRouter().GET("/v1/rk/healthy", healthy)
		entry.GetRouter().GET("/v1/rk/gc", gc)
		entry.GetRouter().GET("/v1/rk/info", info)
		entry.GetRouter().GET("/v1/rk/config", dumpConfig)
		entry.GetRouter().GET("/v1/rk/apis", listApis)
		entry.GetRouter().GET("/v1/rk/sys", sysStats)
		entry.GetRouter().GET("/v1/rk/req", reqStats)
	}

	if entry.enableTV {
		entry.GetRouter().GET("/v1/rk/tv/*item", tv)
	}

	// init server only if port is not zero
	if entry.port != 0 {
		entry.server = &http.Server{
			Addr:    "0.0.0.0:" + strconv.FormatUint(entry.port, 10),
			Handler: entry.router,
		}
	}

	rk_ctx.GlobalAppCtx.AddEntry(entry.GetName(), entry)

	return entry
}

func (entry *GinEntry) GetName() string {
	return entry.name
}

func (entry *GinEntry) GetType() string {
	return entry.entryType
}

func (entry *GinEntry) String() string {
	m := map[string]string{
		"name":         entry.GetName(),
		"type":         entry.GetType(),
		"port":         strconv.FormatUint(entry.GetPort(), 10),
		"interceptors": strconv.Itoa(len(entry.interceptors)),
		"tls":          strconv.FormatBool(entry.IsTlsEnabled()),
		"sw":           strconv.FormatBool(entry.IsSWEnabled()),
	}

	if entry.IsSWEnabled() {
		m["sw_path"] = entry.sw.getPath()
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

func (entry *GinEntry) GetPort() uint64 {
	return entry.port
}

func (entry *GinEntry) GetSWEntry() *swEntry {
	return entry.sw
}

func (entry *GinEntry) IsSWEnabled() bool {
	return entry.sw != nil
}

func (entry *GinEntry) GetTlsEntry() *tlsEntry {
	return entry.tls
}

func (entry *GinEntry) IsTlsEnabled() bool {
	return entry.tls != nil
}

func (entry *GinEntry) GetServer() *http.Server {
	return entry.server
}

func (entry *GinEntry) GetRouter() *gin.Engine {
	return entry.router
}

func (entry *GinEntry) Bootstrap(event rk_query.Event) {
	fields := []zap.Field{
		zap.Uint64("gin_port", entry.port),
		zap.String("gin_name", entry.name),
	}

	if entry.sw != nil {
		fields = append(fields, zap.String("gin_sw_path", entry.sw.getPath()))
	}

	if entry.tls != nil {
		fields = append(fields, zap.Bool("gin_tls", true))
	}

	event.AddFields(fields...)

	go func() {

	}()

	if entry.server != nil {
		// Start server with tls
		if entry.tls != nil {
			entry.logger.Info("starting gin-server", fields...)
			go func(*GinEntry) {
				if err := entry.server.ListenAndServeTLS(entry.tls.GetCertFilePath(), entry.tls.GetKeyFilePath()); err != nil && err != http.ErrServerClosed {
					entry.logger.Error("error while serving gin-listener-tls", fields...)
					shutdownWithError(err)
				}
			}(entry)
		} else {
			entry.logger.Info("starting gin-server", fields...)
			go func(*GinEntry) {
				if err := entry.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					entry.logger.Error("error while serving gin-listener", fields...)
					shutdownWithError(err)
				}
			}(entry)
		}
	}
}

func (entry *GinEntry) Shutdown(event rk_query.Event) {
	fields := []zap.Field{
		zap.Uint64("gin_port", entry.port),
		zap.String("gin_name", entry.name),
	}

	if entry.tls != nil {
		fields = append(fields, zap.Bool("gin_tls", true))
	}

	if entry.sw != nil {
		fields = append(fields, zap.String("gin_sw_path", entry.sw.getPath()))
	}

	event.AddFields(fields...)

	if entry.router != nil && entry.server != nil {
		entry.logger.Info("stopping gin-server", fields...)
		if err := entry.server.Shutdown(context.Background()); err != nil {
			fields = append(fields, zap.Error(err))
			entry.logger.Warn("error occurs while stopping gin-server", fields...)
		}
	}
}

func (entry *GinEntry) Wait(draining time.Duration) {
	sig := <-rk_ctx.GlobalAppCtx.GetShutdownSig()

	helper := rk_query.NewEventHelper(rk_ctx.GlobalAppCtx.GetEventFactory())
	event := helper.Start("rk_app_stop")

	rk_ctx.GlobalAppCtx.GetDefaultLogger().Info("draining", zap.Duration("draining_duration", draining))
	time.Sleep(draining)

	event.AddFields(
		zap.Duration("app_lifetime_nano", time.Since(rk_ctx.GlobalAppCtx.GetStartTime())),
		zap.Time("app_start_time", rk_ctx.GlobalAppCtx.GetStartTime()))

	event.AddPair("signal", sig.String())

	entry.Shutdown(event)

	helper.Finish(event)
}

func readFile(filePath string) []byte {
	if !path.IsAbs(filePath) {
		wd, err := os.Getwd()

		if err != nil {
			shutdownWithError(err)
		}
		filePath = path.Join(wd, filePath)
	}

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		shutdownWithError(err)
	}

	return bytes
}

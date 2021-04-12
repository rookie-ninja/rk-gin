// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	_ "github.com/rookie-ninja/rk-gin/boot/docs"
	"github.com/swaggo/swag"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	swHandlerPrefix = "/swagger/"
	swAssetsPath    = "./assets/swagger-ui/"
)

var (
	swaggerIndexHTML = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>RK Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.35.1/swagger-ui.css" >
    <link rel="icon" type="image/png" href="https://editor.swagger.io/dist/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://editor.swagger.io/dist/favicon-32x32.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }

      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>

  <body>
    <div id="swagger-ui"></div>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.35.1/swagger-ui-bundle.js"> </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.35.1/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
          configUrl: "swagger-config.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIStandalonePreset
          ],
          plugins: [
              SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
      })
      // End Swagger UI call region

      window.ui = ui
    }
  </script>
  </body>
</html>
`
	commonServiceJson, _ = swag.ReadDoc()
	swaggerConfigJson    = ``
	swaggerJsonFiles     = make(map[string]string, 0)
)

// Inner struct used while initializing swagger entry.
type swURLConfig struct {
	URLs []*swURL `json:"urls"`
}

// Inner struct used while initializing swagger entry.
type swURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Bootstrap config of swagger.
// 1: Enabled: Enable swagger.
// 2: Path: Swagger path accessible from restful API.
// 3: JSONPath: The path of where swagger JSON file was located.
// 4: Headers: The headers that would added into each API response.
type BootConfigSW struct {
	Enabled  bool     `yaml:"enabled"`
	Path     string   `yaml:"path"`
	JSONPath string   `yaml:"jsonPath"`
	Headers  []string `yaml:"headers"`
}

// SWEntry implements rkentry.Entry interface.
// 1: Path: Swagger path accessible from restful API.
// 2: JSONPath: The path of where swagger JSON file was located.
// 3: Headers: The headers that would added into each API response.
// 4: Port: The port where swagger would listen to.
type SWEntry struct {
	entryName        string
	entryType        string
	EventLoggerEntry *rkentry.EventLoggerEntry
	ZapLoggerEntry   *rkentry.ZapLoggerEntry
	JSONPath         string
	Path             string
	Headers          map[string]string
	Port             uint64
}

// Swagger entry option.
type SWOption func(*SWEntry)

// Provide port.
func WithPortSW(port uint64) SWOption {
	return func(entry *SWEntry) {
		entry.Port = port
	}
}

// Provide name.
func WithNameSW(name string) SWOption {
	return func(entry *SWEntry) {
		entry.entryName = name
	}
}

// Provide path.
func WithPathSW(path string) SWOption {
	return func(entry *SWEntry) {
		if len(path) < 1 {
			path = "sw"
		}
		entry.Path = path
	}
}

// Provide JSONPath.
func WithJSONPathSW(path string) SWOption {
	return func(entry *SWEntry) {
		entry.JSONPath = path
	}
}

// Provide headers.
func WithHeadersSW(headers map[string]string) SWOption {
	return func(entry *SWEntry) {
		entry.Headers = headers
	}
}

// Provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntrySW(zapLoggerEntry *rkentry.ZapLoggerEntry) SWOption {
	return func(entry *SWEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// Provide rkentry.EventLoggerEntry.
func WithEventLoggerEntrySW(eventLoggerEntry *rkentry.EventLoggerEntry) SWOption {
	return func(entry *SWEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// Create new swagger entry with options.
func NewSWEntry(opts ...SWOption) *SWEntry {
	entry := &SWEntry{
		entryName:        "gin-sw",
		entryType:        "gin-sw",
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		Path:             "sw",
	}

	for i := range opts {
		opts[i](entry)
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if !strings.HasSuffix(entry.Path, "/") {
		entry.Path = entry.Path + "/"
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "gin-sw-" + strconv.FormatUint(entry.Port, 10)
	}

	// init swagger configs
	entry.initSwaggerConfig()

	return entry
}

// Bootstrap swagger entry.
func (entry *SWEntry) Bootstrap(context.Context) {
	// No op
}

// Interrupt swagger entry.
func (entry *SWEntry) Interrupt(context.Context) {
	// No op
}

// Get name of entry.
func (entry *SWEntry) GetName() string {
	return entry.entryName
}

// Get type of entry.
func (entry *SWEntry) GetType() string {
	return entry.entryType
}

// Stringfy swagger entry
func (entry *SWEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
		"json_path":  entry.JSONPath,
		"path":       entry.Path,
		"headers":    entry.Headers,
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Warn("failed to marshal swagger entry to string", zap.Error(err))
		return "{}"
	}

	return string(bytes)
}

// Get gin handler.
func (entry *SWEntry) GinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry.swIndexHandler(c.Writer, c.Request)
	}
}

// Get gin handler for swagger JSON file.
func (entry *SWEntry) GinFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry.swJsonFileHandler(c.Writer, c.Request)
	}
}

// Init swagger config.
// This function do the things bellow:
// 1: List swagger files from entry.JSONPath.
// 2: Read user swagger json files and deduplicate.
// 3: Assign swagger contents into swaggerConfigJson variable
func (entry *SWEntry) initSwaggerConfig() {
	swaggerURLConfig := &swURLConfig{
		URLs: make([]*swURL, 0),
	}

	// 1: Add user API swagger JSON
	entry.listFilesWithSuffix()
	for k := range swaggerJsonFiles {
		swaggerURL := &swURL{
			Name: k,
			URL:  path.Join("/swagger", k),
		}
		entry.appendAndDeduplication(swaggerURLConfig, swaggerURL)
	}

	// 2: Add rk common APIs
	entry.appendAndDeduplication(swaggerURLConfig, &swURL{
		Name: "rk-common",
		URL:  "/swagger/rk_common_service.swagger.json",
	})

	// 3: Marshal to swagger-config.json
	bytes, err := json.Marshal(swaggerURLConfig)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Warn("failed to unmarshal swagger-config.json",
			zap.String("sw_assets_path", swAssetsPath),
			zap.Error(err))
		rkcommon.ShutdownWithError(err)
	}

	swaggerConfigJson = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *SWEntry) listFilesWithSuffix() {
	jsonPath := entry.JSONPath
	suffix := ".json"
	// re-path it with working directory if not absolute path
	if !path.IsAbs(entry.JSONPath) {
		wd, err := os.Getwd()
		if err != nil {
			entry.ZapLoggerEntry.GetLogger().Info("failed to get working directory",
				zap.String("error", err.Error()))
			rkcommon.ShutdownWithError(err)
		}
		jsonPath = path.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil {
		entry.ZapLoggerEntry.GetLogger().Error("failed to list files with suffix",
			zap.String("path", jsonPath),
			zap.String("suffix", suffix),
			zap.String("error", err.Error()))
		rkcommon.ShutdownWithError(err)
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(jsonPath, file.Name()))
			if err != nil {
				entry.ZapLoggerEntry.GetLogger().Info("failed to read file with suffix",
					zap.String("path", path.Join(jsonPath, file.Name())),
					zap.String("suffix", suffix),
					zap.String("error", err.Error()))
				rkcommon.ShutdownWithError(err)
			}

			swaggerJsonFiles[file.Name()] = string(bytes)
		}
	}
}

// Deduplicate based on swagger contents read from JSONPath.
func (entry *SWEntry) appendAndDeduplication(config *swURLConfig, url *swURL) {
	urls := config.URLs

	for i := range urls {
		element := urls[i]

		if element.Name == url.Name {
			return
		}
	}

	config.URLs = append(config.URLs, url)
}

// Swagger file handler.
// Why we need it?
// Because we need to insert common services into swagger.
// Otherwise, we also set headers of <cache-control>:<no-cache> to make sure web browser doesn't cache json config files.
func (entry *SWEntry) swJsonFileHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		http.NotFound(w, r)
		return
	}

	p := strings.TrimPrefix(r.URL.Path, swHandlerPrefix)

	// This is common file
	if p == "rk_common_service.swagger.json" {
		http.ServeContent(w, r, "rk-common", time.Now(), strings.NewReader(commonServiceJson))
		return
	}

	// Set no-cache headers by default
	w.Header().Set("cache-control", "no-cache")

	for k, v := range entry.Headers {
		w.Header().Set(k, v)
	}

	value, ok := swaggerJsonFiles[p]

	if ok {
		http.ServeContent(w, r, p, time.Now(), strings.NewReader(value))
	}
}

// Swagger index handler which specifies <sw> and <sw/swagger-config.json> path.
// Why we need it?
// Because in our swagger HTML, we will try to read swagger config files with path of sw/swagger-config.json.
// As a result, we need to select correct path and return corresponding HTML contents.
func (entry *SWEntry) swIndexHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")
	// This is common file
	if path == "sw" {
		http.ServeContent(w, r, "index.html", time.Now(), strings.NewReader(swaggerIndexHTML))
		return
	} else if path == "sw/swagger-config.json" {
		http.ServeContent(w, r, "swagger-config.json", time.Now(), strings.NewReader(swaggerConfigJson))
		return
	} else {
		return
	}
}

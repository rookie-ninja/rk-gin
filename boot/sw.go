// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	swHandlerPrefix = "/swagger/"
	gwHandlerPrefix = "/"
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
	commonServiceJson = `{
    "swagger": "2.0",
    "info": {
        "description": "This is a common service with rk-gin.",
        "title": "RK Swagger Example",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "paths": {
        "/v1/rk/apis": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "API"
                ],
                "summary": "API",
                "operationId": "5",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/config": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Config"
                ],
                "summary": "Config",
                "operationId": "4",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/gc": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "GC"
                ],
                "summary": "GC",
                "operationId": "2",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/healthy": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Healthy"
                ],
                "summary": "Healthy",
                "operationId": "1",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/info": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Info"
                ],
                "summary": "Info",
                "operationId": "3",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/req": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Request"
                ],
                "summary": "Request Stat",
                "operationId": "7",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/sys": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "System"
                ],
                "summary": "System Stat",
                "operationId": "6",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/rk/tv": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "TV"
                ],
                "summary": "TV",
                "operationId": "8",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`
	swaggerConfigJson = ``
	swaggerJsonFiles  = make(map[string]string, 0)
)

type swURLConfig struct {
	URLs []*swURL `json:"urls"`
}

type swURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type swEntry struct {
	logger   *zap.Logger
	port     uint64
	jsonPath string
	path     string
	headers  map[string]string
}

type swOption func(*swEntry)

func withPort(port uint64) swOption {
	return func(entry *swEntry) {
		entry.port = port
	}
}

func withPath(path string) swOption {
	return func(entry *swEntry) {
		if len(path) < 1 {
			path = "sw"
		}
		entry.path = path
	}
}

func withJsonPath(path string) swOption {
	return func(entry *swEntry) {
		entry.jsonPath = path
	}
}

func withHeaders(headers map[string]string) swOption {
	return func(entry *swEntry) {
		entry.headers = headers
	}
}

func newSWEntry(opts ...swOption) *swEntry {
	entry := &swEntry{
		logger: zap.NewNop(),
	}

	for i := range opts {
		opts[i](entry)
	}

	// Deal with Path
	// add "/" at start and end side if missing
	if !strings.HasPrefix(entry.path, "/") {
		entry.path = "/" + entry.path
	}

	if !strings.HasSuffix(entry.path, "/") {
		entry.path = entry.path + "/"
	}

	// init swagger configs
	entry.initSwaggerConfig()

	return entry
}

func (entry *swEntry) ginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry.swIndexHandler(c.Writer, c.Request)
	}
}

func (entry *swEntry) ginFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry.swJsonFileHandler(c.Writer, c.Request)
	}
}

func (entry *swEntry) getPath() string {
	return entry.path
}

func (entry *swEntry) initSwaggerConfig() {
	// 1: Get swagger-config.json if exists
	swaggerURLConfig := &swURLConfig{
		URLs: make([]*swURL, 0),
	}

	// 2: Add user API swagger JSON
	entry.listFilesWithSuffix()
	for k, _ := range swaggerJsonFiles {
		swaggerURL := &swURL{
			Name: k,
			URL:  path.Join("/swagger", k),
		}
		entry.appendAndDeduplication(swaggerURLConfig, swaggerURL)
	}

	// 3: Add pl-common
	entry.appendAndDeduplication(swaggerURLConfig, &swURL{
		Name: "rk-common",
		URL:  "/swagger/rk_common_service.swagger.json",
	})

	// 4: Marshal to swagger-config.json
	bytes, err := json.Marshal(swaggerURLConfig)
	if err != nil {
		entry.logger.Warn("failed to unmarshal swagger-config.json",
			zap.Uint64("sw_port", entry.port),
			zap.String("sw_assets_path", swAssetsPath),
			zap.Error(err))
		shutdownWithError(err)
	}

	swaggerConfigJson = string(bytes)
}

func (entry *swEntry) listFilesWithSuffix() {
	jsonPath := entry.jsonPath
	suffix := ".json"
	// re-path it with working directory if not absolute path
	if !path.IsAbs(entry.jsonPath) {
		wd, err := os.Getwd()
		if err != nil {
			entry.logger.Info("failed to get working directory",
				zap.String("error", err.Error()))
			shutdownWithError(err)
		}
		jsonPath = path.Join(wd, jsonPath)
	}

	files, err := ioutil.ReadDir(jsonPath)
	if err != nil {
		entry.logger.Error("failed to list files with suffix",
			zap.String("path", jsonPath),
			zap.String("suffix", suffix),
			zap.String("error", err.Error()))
		shutdownWithError(err)
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(jsonPath, file.Name()))
			if err != nil {
				entry.logger.Info("failed to read file with suffix",
					zap.String("path", path.Join(jsonPath, file.Name())),
					zap.String("suffix", suffix),
					zap.String("error", err.Error()))
				shutdownWithError(err)
			}

			swaggerJsonFiles[file.Name()] = string(bytes)
		}
	}
}

func (entry *swEntry) appendAndDeduplication(config *swURLConfig, url *swURL) {
	urls := config.URLs

	for i := range urls {
		element := urls[i]

		if element.Name == url.Name {
			return
		}
	}

	config.URLs = append(config.URLs, url)
}

func (entry *swEntry) swJsonFileHandler(w http.ResponseWriter, r *http.Request) {
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

	for k, v := range entry.headers {
		w.Header().Set(k, v)
	}

	value, ok := swaggerJsonFiles[p]

	if ok {
		http.ServeContent(w, r, p, time.Now(), strings.NewReader(value))
	}
}

func (entry *swEntry) swIndexHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), "/")
	// This is common file
	if path == "sw" {
		http.ServeContent(w, r, "index.html", time.Now(), strings.NewReader(swaggerIndexHTML))
		return
	} else if path == "sw/swagger-config.json" {
		http.ServeContent(w, r, "swagger-config.json", time.Now(), strings.NewReader(swaggerConfigJson))
		return
	} else {

	}
}

<h2 align="center">
  rk-gin
</h2>
<p align="center">
  Inject middlewares & server configuration of <a href="https://grpc.io/docs/languages/go/">gin-gonic/gin</a> from YAML file.
</p>
<p align="center">
  This belongs to <a href="https://github.com/rookie-ninja/rk-boot">rk-boot</a> family. We suggest use this lib with <a href="https://github.com/rookie-ninja/rk-boot">rk-boot</a>.
</p>

<p align="center">
 <a href="https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml"><img src="https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml/badge.svg"></a>
 <a href="https://codecov.io/gh/rookie-ninja/rk-gin"><img src="https://codecov.io/gh/rookie-ninja/rk-gin/branch/master/graph/badge.svg?token=08TCFIIVS0"></a>
 <a href="https://goreportcard.com/badge/github.com/rookie-ninja/rk-gin"><img src="https://goreportcard.com/badge/github.com/rookie-ninja/rk-gin"></a>
 <a href="https://sourcegraph.com/github.com/rookie-ninja/rk-gin?badge"><img src="https://sourcegraph.com/github.com/rookie-ninja/rk-gin/-/badge.svg"></a>
 <a href="https://godoc.org/github.com/rookie-ninja/rk-gin"><img src="https://godoc.org/github.com/rookie-ninja/rk-gin?status.svg"></a>
 <a href="https://github.com/rookie-ninja/rk-gin/releases"><img src="https://img.shields.io/github/release/rookie-ninja/rk-gin.svg?style=flat-square"></a>
 <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg"></a>
<p>

<div id="badges" align="center">
  <a href="https://rkdev.info">
    <img src="https://img.shields.io/badge/Official Site-blue?logo=mdbook&logoColor=white&style=for-the-badge" alt="Docs Badge"/>
  </a>
  <a href="https://rk-syz1767.slack.com/rk-boot">
    <img src="https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white" alt="Docs Badge"/>
  </a>
</div>

## Architecture
![image](docs/img/gin-arch.png)

## Quick Start
In the bellow example, we will start microservice with bellow functionality and middlewares enabled via YAML.

- [gin-gonic/gin](https://github.com/gin-gonic/gin) server
- Swagger UI
- Docs
- CommonService
- Prometheus Metrics (middleware)
- Logging (middleware)
- Meta (middleware)

Please refer example at [example/boot/simple](example/boot/simple).

### Installation
```shell
go get github.com/rookie-ninja/rk-gin/v2
```

### 1.Create boot.yaml
- [boot.yaml](example/boot/simple/boot.yaml)

<details>
<summary>show</summary>

```yaml
---
gin:
  - name: greeter                     # Required
    port: 8080                        # Required
    enabled: true                     # Required
    commonService:                    # Optional
      enabled: true                   # Optional, default: false
    sw:                               # Optional
      enabled: true                   # Optional, default: false
    docs:                             # Optional
      enabled: true                   # Optional, default: false
    prom:
      enabled: true                   # Optional, default: false
    middleware:
      logging:
        enabled: true
      prom:
        enabled: true
      meta:
        enabled: true
```

</details>

### 2.Create main.go
- [main.go](example/boot/simple/main.go)

<details>
<summary>show</summary>

```go
// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-gin/v2/boot"
	"net/http"
)

// How to use embed.FS for:
//
// - boot.yaml
// - rkentry.DocsEntryType
// - rkentry.SWEntryType
// - rkentry.StaticFileHandlerEntryType
// - rkentry.CertEntry
//
// If we use embed.FS, then we only need one single binary file while packing.
// We suggest use embed.FS to pack swagger local file since rk-entry would use os.Getwd() to look for files
// if relative path was provided.
//
//go:embed docs
var docsFS embed.FS

func init() {
	rkentry.GlobalAppCtx.AddEmbedFS(rkentry.SWEntryType, "greeter", &docsFS)
}

//go:embed boot.yaml
var boot []byte

// @title RK Swagger for Gin
// @version 1.0
// @description This is a greeter service with rk-boot.
func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntryYAML(boot)

	// Get GinEntry
	ginEntry := res["greeter"].(*rkgin.GinEntry)
	ginEntry.Router.GET("/v1/greeter", Greeter)

	// Bootstrap gin entry
	ginEntry.Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	ginEntry.Interrupt(context.Background())
}

// Greeter handler
// @Summary Greeter service
// @Id 1
// @version 1.0
// @produce application/json
// @Param name query string true "Input name"
// @Success 200 {object} GreeterResponse
// @Router /v1/greeter [get]
func Greeter(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

type GreeterResponse struct {
	Message string
}
```

</details>

### 3.Start server

```go
$ go run main.go
```

### 4.Validation

<details>
<summary>show</summary>

#### 4.1 Gin server
Try to test Gin Service with [curl](https://curl.se/)

```shell script
# Curl to common service
$ curl localhost:8080/rk/v1/ready
{
  "ready": true
}

$ curl localhost:8080/rk/v1/alive
{
  "alive": true
}
```

#### 4.2 Swagger UI
Please refer **sw** section at [Full YAML](#full-yaml).

By default, we could access swagger UI at [http://localhost:8080/sw](http://localhost:8080/sw)

![sw](docs/img/simple-sw.png)

#### 4.3 Docs UI
Please refer **docs** section at [Full YAML](#full-yaml).

By default, we could access docs UI at [http://localhost:8080/docs](http://localhost:8080/docs)

![docs](docs/img/simple-docs.png)

#### 4.4 Prometheus Metrics
Please refer **middleware.prom** section at [Full YAML](#full-yaml).

By default, we could access prometheus client at [http://localhost:8080/metrics](http://localhost:8080/metrics)
- http://localhost:8080/metrics

![prom](docs/img/simple-prom.png)

#### 4.5 Logging
Please refer **middleware.logging** section at [Full YAML](#full-yaml).

By default, we enable zap logger and event logger with encoding type of [console]. Encoding type of [json] and [flatten] is also supported.

```shell script
2021-12-28T02:14:48.303+0800    INFO    boot/gin_entry.go:920   Bootstrap ginEntry      {"eventId": "65b03dbc-c10e-4998-8d49-26775dafc78b", "entryName": "greeter"}
------------------------------------------------------------------------
endTime=2021-12-28T02:14:48.305036+08:00
startTime=2021-12-28T02:14:48.30306+08:00
elapsedNano=1977443
timezone=CST
ids={"eventId":"65b03dbc-c10e-4998-8d49-26775dafc78b"}
app={"appName":"rk","appVersion":"","entryName":"greeter","entryType":"GinEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"commonServiceEnabled":true,"commonServicePathPrefix":"/rk/v1/","entryName":"greeter","entryPort":8080,"entryType":"GinEntry","promEnabled":true,"promPath":"/metrics","promPort":8080,"swEnabled":true,"swPath":"/sw/","tvEnabled":true,"tvPath":"/rk/v1/tv/"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=Bootstrap
resCode=OK
eventStatus=Ended
EOE
```

#### 4.6 Meta
Please refer **meta** section at [Full YAML](#full-yaml).

By default, we will send back some metadata to client including gateway with headers.

```shell script
$ curl -vs localhost:8080/rk/v1/ready
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /rk/v1/healthy HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.64.1
> Accept: */*
> 
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Request-Id: f3f0212e-5d99-4851-ae79-ea88818f0ed6
< X-Rk-App-Name: rk
< X-Rk-App-Unix-Time: 2021-12-28T02:20:48.207716+08:00
< X-Rk-Received-Time: 2021-12-28T02:20:48.207716+08:00
< Date: Mon, 27 Dec 2021 18:20:48 GMT
< Content-Length: 16
< 
* Connection #0 to host localhost left intact
{"ready":true}
```

#### 4.7 Send request
We registered /v1/greeter API in [gin-gonic/gin](https://github.com/gin-gonic/gin) server and let's validate it!

```shell script
$ curl -vs "localhost:8080/v1/greeter?name=rk-dev"
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /v1/greeter?name=rk-dev HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.64.1
> Accept: */*
> 
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Request-Id: a96ab531-e28f-47ca-a082-fc3f8ef14187
< X-Rk-App-Name: rk
< X-Rk-App-Unix-Time: 2021-12-28T02:22:03.289469+08:00
< X-Rk-Received-Time: 2021-12-28T02:22:03.289469+08:00
< Date: Mon, 27 Dec 2021 18:22:03 GMT
< Content-Length: 27
< 
* Connection #0 to host localhost left intact
{"Message":"Hello rk-dev!"}
```

#### 4.8 RPC logs
Bellow logs would be printed in stdout.

```
------------------------------------------------------------------------
endTime=2021-12-28T02:22:03.289585+08:00
startTime=2021-12-28T02:22:03.289457+08:00
elapsedNano=128210
timezone=CST
ids={"eventId":"a96ab531-e28f-47ca-a082-fc3f8ef14187","requestId":"a96ab531-e28f-47ca-a082-fc3f8ef14187"}
app={"appName":"rk","appVersion":"","entryName":"greeter","entryType":"GinEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:54028
operation=/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

#### 4.9 RPC prometheus metrics
Prometheus client will automatically register into [gin-gonic/gin](https://github.com/gin-gonic/gin) instance at /metrics.

Access [http://localhost:8080/metrics](http://localhost:8080/metrics)

![image](docs/img/prom-inter.png)

</details>



## Supported features
**User can enable anyone of those as needed! No mandatory binding!**

| Instance          | Description                                                                                                   |
|-------------------|---------------------------------------------------------------------------------------------------------------|
| gin.Router        | Compatible with original [gin-gonic/gin](https://github.com/gin-gonic/gin) service functionalities            |
| Config            | Configure [spf13/viper](https://github.com/spf13/viper) as config instance and reference it from YAML         |
| Logger            | Configure [uber-go/zap](https://github.com/uber-go/zap) logger configuration and reference it from YAML       |
| Event             | Configure logging of RPC with [rk-query](https://github.com/rookie-ninja/rk-query) and reference it from YAML |
| Cert              | Fetch TLS/SSL certificates start microservice.                                                                |
| Prometheus        | Start prometheus client at client side and push metrics to pushgateway as needed.                             |
| Swagger           | Builtin swagger UI handler.                                                                                   |
| Docs              | Builtin [RapiDoc](https://github.com/mrin9/RapiDoc) instance which can be used to replace swagger and RK TV.  |
| CommonService     | List of common APIs.                                                                                          |
| StaticFileHandler | A Web UI shows files could be downloaded from server, currently support source of local and embed.FS.         |
| PProf             | PProf web UI.                                                                                                 |

## Supported middlewares
All middlewares could be configured via YAML or Code.

**User can enable anyone of those as needed! No mandatory binding!**

| Middleware | Description                                                                                                                                           |
|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| Prom       | Collect RPC metrics and export to [prometheus](https://github.com/prometheus/client_golang) client.                                                   |
| Logging    | Log every RPC requests as event with [rk-query](https://github.com/rookie-ninja/rk-query).                                                            |
| Trace      | Collect RPC trace and export it to stdout, file or jaeger with [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go). |
| Panic      | Recover from panic for RPC requests and log it.                                                                                                       |
| Meta       | Send micsro service metadata as header to client.                                                                                                     |
| Auth       | Support [Basic Auth] and [API Key] authorization types.                                                                                               |
| RateLimit  | Limiting RPC rate globally or per path.                                                                                                               |
| Timeout    | Timing out request by configuration.                                                                                                                  |
| Gzip       | Compress and Decompress message body based on request header with gzip format .                                                                       |
| CORS       | Server side CORS validation.                                                                                                                          |
| JWT        | Server side JWT validation.                                                                                                                           |
| Secure     | Server side secure validation.                                                                                                                        |
| CSRF       | Server side csrf validation.                                                                                                                          |

## YAML Options
User can start multiple [gin-gonic/gin](https://github.com/gin-gonic/gin) instances at the same time. Please make sure use different port and name.

<details>
<summary>show</summary>

```yaml
---
#app:
#  name: my-app                                            # Optional, default: "rk-app"
#  version: "v1.0.0"                                       # Optional, default: "v0.0.0"
#  description: "this is description"                      # Optional, default: ""
#  keywords: ["rk", "golang"]                              # Optional, default: []
#  homeUrl: "http://example.com"                           # Optional, default: ""
#  docsUrl: ["http://example.com"]                         # Optional, default: []
#  maintainers: ["rk-dev"]                                 # Optional, default: []
#logger:
#  - name: my-logger                                       # Required
#    description: "Description of entry"                   # Optional
#    domain: "*"                                           # Optional, default: "*"
#    default: false                                        # Optional, default: false, use as default logger entry
#    zap:                                                  # Optional
#      level: info                                         # Optional, default: info
#      development: true                                   # Optional, default: true
#      disableCaller: false                                # Optional, default: false
#      disableStacktrace: true                             # Optional, default: true
#      encoding: console                                   # Optional, default: console
#      outputPaths: ["stdout"]                             # Optional, default: [stdout]
#      errorOutputPaths: ["stderr"]                        # Optional, default: [stderr]
#      encoderConfig:                                      # Optional
#        timeKey: "ts"                                     # Optional, default: ts
#        levelKey: "level"                                 # Optional, default: level
#        nameKey: "logger"                                 # Optional, default: logger
#        callerKey: "caller"                               # Optional, default: caller
#        messageKey: "msg"                                 # Optional, default: msg
#        stacktraceKey: "stacktrace"                       # Optional, default: stacktrace
#        skipLineEnding: false                             # Optional, default: false
#        lineEnding: "\n"                                  # Optional, default: \n
#        consoleSeparator: "\t"                            # Optional, default: \t
#      sampling:                                           # Optional, default: nil
#        initial: 0                                        # Optional, default: 0
#        thereafter: 0                                     # Optional, default: 0
#      initialFields:                                      # Optional, default: empty map
#        key: value
#    lumberjack:                                           # Optional, default: nil
#      filename:
#      maxsize: 1024                                       # Optional, suggested: 1024 (MB)
#      maxage: 7                                           # Optional, suggested: 7 (day)
#      maxbackups: 3                                       # Optional, suggested: 3 (day)
#      localtime: true                                     # Optional, suggested: true
#      compress: true                                      # Optional, suggested: true
#    loki:
#      enabled: true                                       # Optional, default: false
#      addr: localhost:3100                                # Optional, default: localhost:3100
#      path: /loki/api/v1/push                             # Optional, default: /loki/api/v1/push
#      username: ""                                        # Optional, default: ""
#      password: ""                                        # Optional, default: ""
#      maxBatchWaitMs: 3000                                # Optional, default: 3000
#      maxBatchSize: 1000                                  # Optional, default: 1000
#      insecureSkipVerify: false                           # Optional, default: false
#      labels:                                             # Optional, default: empty map
#        my_label_key: my_label_value
#event:
#  - name: my-event                                        # Required
#    description: "Description of entry"                   # Optional
#    domain: "*"                                           # Optional, default: "*"
#    default: false                                        # Optional, default: false, use as default event entry
#    encoding: console                                     # Optional, default: console
#    outputPaths: ["stdout"]                               # Optional, default: [stdout]
#    lumberjack:                                           # Optional, default: nil
#      filename:
#      maxsize: 1024                                       # Optional, suggested: 1024 (MB)
#      maxage: 7                                           # Optional, suggested: 7 (day)
#      maxbackups: 3                                       # Optional, suggested: 3 (day)
#      localtime: true                                     # Optional, suggested: true
#      compress: true                                      # Optional, suggested: true
#    loki:
#      enabled: true                                       # Optional, default: false
#      addr: localhost:3100                                # Optional, default: localhost:3100
#      path: /loki/api/v1/push                             # Optional, default: /loki/api/v1/push
#      username: ""                                        # Optional, default: ""
#      password: ""                                        # Optional, default: ""
#      maxBatchWaitMs: 3000                                # Optional, default: 3000
#      maxBatchSize: 1000                                  # Optional, default: 1000
#      insecureSkipVerify: false                           # Optional, default: false
#      labels:                                             # Optional, default: empty map
#        my_label_key: my_label_value
#cert:
#  - name: my-cert                                         # Required
#    description: "Description of entry"                   # Optional, default: ""
#    domain: "*"                                           # Optional, default: "*"
#    caPath: "certs/ca.pem"                                # Optional, default: ""
#    certPemPath: "certs/server-cert.pem"                  # Optional, default: ""
#    keyPemPath: "certs/server-key.pem"                    # Optional, default: ""
#config:
#  - name: my-config                                       # Required
#    description: "Description of entry"                   # Optional, default: ""
#    domain: "*"                                           # Optional, default: "*"
#    path: "config/config.yaml"                            # Optional
#    envPrefix: ""                                         # Optional, default: ""
#    content:                                              # Optional, defualt: empty map
#      key: value
gin:
  - name: greeter                                          # Required
    port: 8080                                             # Required
    enabled: true                                          # Required
#    description: "greeter server"                         # Optional, default: ""
#    certEntry: my-cert                                    # Optional, default: "", reference of cert entry declared above
#    loggerEntry: my-logger                                # Optional, default: "", reference of cert entry declared above, STDOUT will be used if missing
#    eventEntry: my-event                                  # Optional, default: "", reference of cert entry declared above, STDOUT will be used if missing
#    sw:
#      enabled: true                                       # Optional, default: false
#      path: "sw"                                          # Optional, default: "sw"
#      jsonPath: ""                                        # Optional
#      headers: ["sw:rk"]                                  # Optional, default: []
#    docs:
#      enabled: true                                       # Optional, default: false
#      path: "docs"                                        # Optional, default: "docs"
#      specPath: ""                                        # Optional
#      headers: ["sw:rk"]                                  # Optional, default: []
#      style:                                              # Optional
#        theme: "light"                                    # Optional, default: "light"
#      debug: false                                        # Optional, default: false
#    commonService:
#      enabled: true                                       # Optional, default: false
#      pathPrefix: ""                                      # Optional, default: "/rk/v1/"
#    static:
#      enabled: true                                       # Optional, default: false
#      path: "/static"                                     # Optional, default: /static
#      sourceType: local                                   # Optional, options: local, embed.FS can be used either, need to specify in code
#      sourcePath: "."                                     # Optional, full path of source directory
#    pprof:
#      enabled: true                                       # Optional, default: false
#      path: "/pprof"                                      # Optional, default: /pprof
#    prom:
#      enabled: true                                       # Optional, default: false
#      path: ""                                            # Optional, default: "/metrics"
#      pusher:
#        enabled: false                                    # Optional, default: false
#        jobName: "greeter-pusher"                         # Required
#        remoteAddress: "localhost:9091"                   # Required
#        basicAuth: "user:pass"                            # Optional, default: ""
#        intervalMs: 10000                                 # Optional, default: 1000
#        certEntry: my-cert                                # Optional, default: "", reference of cert entry declared above
#    middleware:
#      ignore: [""]                                        # Optional, default: []
#      errorModel: google                                  # Optional, default: google, [amazon, google] are supported options
#      logging:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        loggerEncoding: "console"                         # Optional, default: "console"
#        loggerOutputPaths: ["logs/app.log"]               # Optional, default: ["stdout"]
#        eventEncoding: "console"                          # Optional, default: "console"
#        eventOutputPaths: ["logs/event.log"]              # Optional, default: ["stdout"]
#      prom:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#      auth:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        basic:
#          - "user:pass"                                   # Optional, default: []
#        apiKey:
#          - "keys"                                        # Optional, default: []
#      meta:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        prefix: "rk"                                      # Optional, default: "rk"
#      trace:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        exporter:                                         # Optional, default will create a stdout exporter
#          file:
#            enabled: true                                 # Optional, default: false
#            outputPath: "logs/trace.log"                  # Optional, default: stdout
#          jaeger:
#            agent:
#              enabled: false                              # Optional, default: false
#              host: ""                                    # Optional, default: localhost
#              port: 0                                     # Optional, default: 6831
#            collector:
#              enabled: true                               # Optional, default: false
#              endpoint: ""                                # Optional, default: http://localhost:14268/api/traces
#              username: ""                                # Optional, default: ""
#              password: ""                                # Optional, default: ""
#      rateLimit:
#        enabled: false                                    # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        algorithm: "leakyBucket"                          # Optional, default: "tokenBucket"
#        reqPerSec: 100                                    # Optional, default: 1000000
#        paths:
#          - path: "/rk/v1/healthy"                        # Optional, default: ""
#            reqPerSec: 0                                  # Optional, default: 1000000
#      timeout:
#        enabled: false                                    # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        timeoutMs: 5000                                   # Optional, default: 5000
#        paths:
#          - path: "/rk/v1/healthy"                        # Optional, default: ""
#            timeoutMs: 1000                               # Optional, default: 5000
#      jwt:
#        enabled: true                                     # Optional, default: false
#        ignore: [ "" ]                                    # Optional, default: []
#        signerEntry: ""                                   # Optional, default: ""
#        skipVerify: false                                 # Optional, default: false
#        symmetric:                                        # Optional
#          algorithm: ""                                   # Required, default: ""
#          token: ""                                       # Optional, default: ""
#          tokenPath: ""                                   # Optional, default: ""
#        asymmetric:                                       # Optional
#          algorithm: ""                                   # Required, default: ""
#          privateKey: ""                                  # Optional, default: ""
#          privateKeyPath: ""                              # Optional, default: ""
#          publicKey: ""                                   # Optional, default: ""
#          publicKeyPath: ""                               # Optional, default: ""
#        tokenLookup: "header:<name>"                      # Optional, default: "header:Authorization"
#        authScheme: "Bearer"                              # Optional, default: "Bearer"
#      secure:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        xssProtection: ""                                 # Optional, default: "1; mode=block"
#        contentTypeNosniff: ""                            # Optional, default: nosniff
#        xFrameOptions: ""                                 # Optional, default: SAMEORIGIN
#        hstsMaxAge: 0                                     # Optional, default: 0
#        hstsExcludeSubdomains: false                      # Optional, default: false
#        hstsPreloadEnabled: false                         # Optional, default: false
#        contentSecurityPolicy: ""                         # Optional, default: ""
#        cspReportOnly: false                              # Optional, default: false
#        referrerPolicy: ""                                # Optional, default: ""
#      csrf:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        tokenLength: 32                                   # Optional, default: 32
#        tokenLookup: "header:X-CSRF-Token"                # Optional, default: "header:X-CSRF-Token"
#        cookieName: "_csrf"                               # Optional, default: _csrf
#        cookieDomain: ""                                  # Optional, default: ""
#        cookiePath: ""                                    # Optional, default: ""
#        cookieMaxAge: 86400                               # Optional, default: 86400
#        cookieHttpOnly: false                             # Optional, default: false
#        cookieSameSite: "default"                         # Optional, default: "default", options: lax, strict, none, default
#      gzip:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        level: bestSpeed                                  # Optional, options: [noCompression, bestSpeed， bestCompression, defaultCompression, huffmanOnly]
#      cors:
#        enabled: true                                     # Optional, default: false
#        ignore: [""]                                      # Optional, default: []
#        allowOrigins:                                     # Optional, default: []
#          - "http://localhost:*"                          # Optional, default: *
#        allowCredentials: false                           # Optional, default: false
#        allowHeaders: []                                  # Optional, default: []
#        allowMethods: []                                  # Optional, default: []
#        exposeHeaders: []                                 # Optional, default: []
#        maxAge: 0                                         # Optional, default: 0
```

</details>

## Development Status: Stable

## Build instruction
Simply run make all to validate your changes. Or run codes in example/ folder.

- make all

If files in boot/assets have been modified, then we need to run it.

## Test instruction
Run unit test with **make test** command.

github workflow will automatically run unit test and golangci-lint for testing and lint validation.

## Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).


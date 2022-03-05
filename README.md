# rk-gin
[![build](https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-gin/branch/master/graph/badge.svg?token=S0B7KTMIHW)](https://codecov.io/gh/rookie-ninja/rk-gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-gin)](https://goreportcard.com/report/github.com/rookie-ninja/rk-gin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Middleware & bootstrapper designed for [gin-gonic/gin](https://github.com/gin-gonic/gin) web framework. [Documentation](https://rkdev.info/docs/bootstrapper/user-guide/gin-golang/).

This belongs to [rk-boot](https://github.com/rookie-ninja/rk-boot) family. We suggest use this lib from [rk-boot](https://github.com/rookie-ninja/rk-boot).

![image](docs/img/boot-arch.png)

## Architecture
![image](docs/img/gin-arch.png)

## Supported bootstrap
| Bootstrap  | Description                                                                    |
|------------|--------------------------------------------------------------------------------|
| YAML based | Start [gin-gonic/gin](https://github.com/gin-gonic/gin) microservice from YAML |
| Code based | Start [gin-gonic/gin](https://github.com/gin-gonic/gin) microservice from code |

## Supported instances
All instances could be configured via YAML or Code.

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
| StaticFileHandler | A Web UI shows files could be downloaded from server, currently support source of local and pkger.            |

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

## Installation
`go get github.com/rookie-ninja/rk-gin/v2`

## Quick Start
In the bellow example, we will start microservice with bellow functionality and middlewares enabled via YAML.

- [gin-gonic/gin](https://github.com/gin-gonic/gin) server
- Swagger UI
- CommonService
- Prometheus Metrics (middleware)
- Logging (middleware)
- Meta (middleware)

Please refer example at [example/boot/simple](example/boot/simple).

### 1.Create boot.yaml
- [boot.yaml](example/boot/simple/boot.yaml)

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

### 2.Create main.go
- [main.go](example/boot/simple/main.go)

```go
// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
  "context"
  _ "embed"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/rookie-ninja/rk-entry/v2/entry"
  "github.com/rookie-ninja/rk-gin/v2/boot"
  "net/http"
)

//go:embed boot.yaml
var boot []byte

////go:embed docs
//var docsFS embed.FS
//
////go:embed docs
//var staticFS embed.FS

func init() {
  //rkentry.GlobalAppCtx.AddEmbedFS(rkentry.DocsEntryType, "greeter", &docsFS)
  //rkentry.GlobalAppCtx.AddEmbedFS(rkentry.SWEntryType, "greeter", &docsFS)
  //rkentry.GlobalAppCtx.AddEmbedFS(rkentry.StaticFileHandlerEntryType, "greeter", &staticFS)
}

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

func Greeter(ctx *gin.Context) {
  ctx.JSON(http.StatusOK, &GreeterResponse{
    Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
  })
}

type GreeterResponse struct {
  Message string
}
```

### 3.Start server

```go
$ go run main.go
```

### 4.Validation
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

By default, we enable zap logger and event logger with encoding type of [console]. Encoding type of [json] is also supported.

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
$ curl -vs localhost:8080/rk/v1/healthy
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
{"healthy":true}
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

## YAML Options
User can start multiple [gin-gonic/gin](https://github.com/gin-gonic/gin) instances at the same time. Please make sure use different port and name.

### Gin
| name            | description                                                                                                            | type    | default value           |
|-----------------|------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| gin.name        | Required, The name of gin server                                                                                       | string  | N/A                     |
| gin.port        | Required, The port of gin server                                                                                       | integer | nil, server won't start |
| gin.enabled     | Optional, Enable Gin entry or not                                                                                      | bool    | false                   |
| gin.description | Optional, Description of gin entry.                                                                                    | string  | ""                      |
| gin.certEntry   | Optional, Reference of certEntry declared in [cert entry](https://github.com/rookie-ninja/rk-entry#certentry)          | string  | ""                      |
| gin.loggerEntry | Optional, Reference of loggerEntry declared in [zapLoggerEntry](https://github.com/rookie-ninja/rk-entry#loggerentry)  | string  | ""                      |
| gin.eventEntry  | Optional, Reference of eventLEntry declared in [eventLoggerEntry](https://github.com/rookie-ninja/rk-entry#evententry) | string  | ""                      |

### CommonService
| Path         | Description                       |
|--------------|-----------------------------------|
| /rk/v1/gc    | Trigger GC                        |
| /rk/v1/ready | Get application readiness status. |
| /rk/v1/alive | Get application aliveness status. |
| /rk/v1/info  | Get application and process info. |

| name                         | description                             | type    | default value |
|------------------------------|-----------------------------------------|---------|---------------|
| gin.commonService.enabled    | Optional, Enable builtin common service | boolean | false         |
| gin.commonService.pathPrefix | Optional, Provide path prefix           | string  | /rk/v1        |

### Swagger
| name            | description                                                        | type     | default value |
|-----------------|--------------------------------------------------------------------|----------|---------------|
| gin.sw.enabled  | Optional, Enable swagger service over gin server                   | boolean  | false         |
| gin.sw.path     | Optional, The path access swagger service from web                 | string   | /sw           |
| gin.sw.jsonPath | Optional, Where the swagger.json files are stored locally          | string   | ""            |
| gin.sw.headers  | Optional, Headers would be sent to caller as scheme of [key:value] | []string | []            |

### Docs (RapiDoc)
| name                 | description                                                                            | type     | default value |
|----------------------|----------------------------------------------------------------------------------------|----------|---------------|
| gin.docs.enabled     | Optional, Enable RapiDoc service over gin server                                       | boolean  | false         |
| gin.docs.path        | Optional, The path access docs service from web                                        | string   | /docs         |
| gin.docs.jsonPath    | Optional, Where the swagger.json or open API files are stored locally                  | string   | ""            |
| gin.docs.headers     | Optional, Headers would be sent to caller as scheme of [key:value]                     | []string | []            |
| gin.docs.style.theme | Optional, light and dark are supported options                                         | string   | []            |
| gin.docs.debug       | Optional, Enable debugging mode in RapiDoc which can be used as the same as Swagger UI | boolean  | false         |

### Prom Client
| name                          | description                                                                        | type    | default value |
|-------------------------------|------------------------------------------------------------------------------------|---------|---------------|
| gin.prom.enabled              | Optional, Enable prometheus                                                        | boolean | false         |
| gin.prom.path                 | Optional, Path of prometheus                                                       | string  | /metrics      |
| gin.prom.pusher.enabled       | Optional, Enable prometheus pusher                                                 | bool    | false         |
| gin.prom.pusher.jobName       | Optional, Job name would be attached as label while pushing to remote pushgateway  | string  | ""            |
| gin.prom.pusher.remoteAddress | Optional, PushGateWay address, could be form of http://x.x.x.x or x.x.x.x          | string  | ""            |
| gin.prom.pusher.intervalMs    | Optional, Push interval in milliseconds                                            | string  | 1000          |
| gin.prom.pusher.basicAuth     | Optional, Basic auth used to interact with remote pushgateway, form of [user:pass] | string  | ""            |
| gin.prom.pusher.certEntry     | Optional, Reference of rkentry.CertEntry                                           | string  | ""            |

### Static file handler
| name                  | description                                | type    | default value |
|-----------------------|--------------------------------------------|---------|---------------|
| gin.static.enabled    | Optional, Enable static file handler       | boolean | false         |
| gin.static.path       | Optional, path of static file handler      | string  | /static       |
| gin.static.sourceType | Required, local and embed.FS are supported | string  | ""            |
| gin.static.sourcePath | Required, full path of source directory    | string  | ""            |

- About embed.FS
User has to set embedFS before Bootstrap() function as bellow:
- 
```go
//go:embed /*
var staticFS embed.FS

rkentry.GlobalAppCtx.AddEmbedFS(rkentry.StaticFileHandlerEntryType, "greeter", &staticFS)
```

### Middlewares
| name                  | description                                            | type     | default value |
|-----------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.ignore | The paths of prefix that will be ignored by middleware | []string | []            |

#### Logging
| name                                     | description                                            | type     | default value |
|------------------------------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.logging.enabled           | Enable log middleware                                  | boolean  | false         |
| gin.middleware.logging.ignore            | The paths of prefix that will be ignored by middleware | []string | []            |
| gin.middleware.logging.loggerEncoding    | json or console or flatten                             | string   | console       |
| gin.middleware.logging.loggerOutputPaths | Output paths                                           | []string | stdout        |
| gin.middleware.logging.eventEncoding     | json or console or flatten                             | string   | console       |
| gin.middleware.logging.eventOutputPaths  | Output paths                                           | []string | false         |

We will log two types of log for every RPC call.
- Logger

Contains user printed logging with requestId or traceId.

- Event

Contains per RPC metadata, response information, environment information and etc.

| Field       | Description                                                                                                                                                                                                                                                                                                        |
|-------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| endTime     | As name described                                                                                                                                                                                                                                                                                                  |
| startTime   | As name described                                                                                                                                                                                                                                                                                                  |
| elapsedNano | Elapsed time for RPC in nanoseconds                                                                                                                                                                                                                                                                                |
| timezone    | As name described                                                                                                                                                                                                                                                                                                  |
| ids         | Contains three different ids(eventId, requestId and traceId). If meta middleware was enabled or event.SetRequestId() was called by user, then requestId would be attached. eventId would be the same as requestId if meta middleware was enabled. If trace middleware was enabled, then traceId would be attached. |
| app         | Contains [appName, appVersion](https://github.com/rookie-ninja/rk-entry#appinfoentry), entryName, entryType.                                                                                                                                                                                                       |
| env         | Contains arch, az, domain, hostname, localIP, os, realm, region. realm, region, az, domain were retrieved from environment variable named as REALM, REGION, AZ and DOMAIN. "*" means empty environment variable.                                                                                                   |
| payloads    | Contains RPC related metadata                                                                                                                                                                                                                                                                                      |
| error       | Contains errors if occur                                                                                                                                                                                                                                                                                           |
| counters    | Set by calling event.SetCounter() by user.                                                                                                                                                                                                                                                                         |
| pairs       | Set by calling event.AddPair() by user.                                                                                                                                                                                                                                                                            |
| timing      | Set by calling event.StartTimer() and event.EndTimer() by user.                                                                                                                                                                                                                                                    |
| remoteAddr  | As name described                                                                                                                                                                                                                                                                                                  |
| operation   | RPC method name                                                                                                                                                                                                                                                                                                    |
| resCode     | Response code of RPC                                                                                                                                                                                                                                                                                               |
| eventStatus | Ended or InProgress                                                                                                                                                                                                                                                                                                |

- example

```shell script
------------------------------------------------------------------------
endTime=2021-06-25T01:30:45.144023+08:00
startTime=2021-06-25T01:30:45.143767+08:00
elapsedNano=255948
timezone=CST
ids={"eventId":"3332e575-43d8-4bfe-84dd-45b5fc5fb104","requestId":"3332e575-43d8-4bfe-84dd-45b5fc5fb104","traceId":"65b9aa7a9705268bba492fdf4a0e5652"}
app={"appName":"rk-gin","appVersion":"master-xxx","entryName":"greeter","entryType":"GinEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/healthy","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:60718
operation=/rk/v1/healthy
resCode=200
eventStatus=Ended
EOE
```

#### Prometheus
| name                        | description                                            | type     | default value |
|-----------------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.prom.enabled | Enable metrics middleware                              | boolean  | false         |
| gin.middleware.prom.ignore  | The paths of prefix that will be ignored by middleware | []string | []            |

#### Auth
Enable the server side auth. codes.Unauthenticated would be returned to client if not authorized with user defined credential.

| name                        | description                                            | type     | default value |
|-----------------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.auth.enabled | Enable auth middleware                                 | boolean  | false         |
| gin.middleware.auth.ignore  | The paths of prefix that will be ignored by middleware | []string | []            |
| gin.middleware.auth.basic   | Basic auth credentials as scheme of <user:pass>        | []string | []            |
| gin.middleware.auth.apiKey  | API key auth                                           | []string | []            |

#### Meta
Send application metadata as header to client.

| name                        | description                                            | type     | default value |
|-----------------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.meta.enabled | Enable meta middleware                                 | boolean  | false         |
| gin.middleware.meta.ignore  | The paths of prefix that will be ignored by middleware | []string | []            |
| gin.middleware.meta.prefix  | Header key was formed as X-<Prefix>-XXX                | string   | RK            |

#### Trace
| name                                                    | description                                            | type     | default value                    |
|---------------------------------------------------------|--------------------------------------------------------|----------|----------------------------------|
| gin.middleware.trace.enabled                            | Enable tracing middleware                              | boolean  | false                            |
| gin.middleware.trace.ignore                             | The paths of prefix that will be ignored by middleware | []string | []                               |
| gin.middleware.trace.exporter.file.enabled              | Enable file exporter                                   | boolean  | false                            |
| gin.middleware.trace.exporter.file.outputPath           | Export tracing info to files                           | string   | stdout                           |
| gin.middleware.trace.exporter.jaeger.agent.enabled      | Export tracing info to jaeger agent                    | boolean  | false                            |
| gin.middleware.trace.exporter.jaeger.agent.host         | As name described                                      | string   | localhost                        |
| gin.middleware.trace.exporter.jaeger.agent.port         | As name described                                      | int      | 6831                             |
| gin.middleware.trace.exporter.jaeger.collector.enabled  | Export tracing info to jaeger collector                | boolean  | false                            |
| gin.middleware.trace.exporter.jaeger.collector.endpoint | As name described                                      | string   | http://localhost:16368/api/trace |
| gin.middleware.trace.exporter.jaeger.collector.username | As name described                                      | string   | ""                               |
| gin.middleware.trace.exporter.jaeger.collector.password | As name described                                      | string   | ""                               |

#### RateLimit
| name                                     | description                                                          | type     | default value |
|------------------------------------------|----------------------------------------------------------------------|----------|---------------|
| gin.middleware.rateLimit.enabled         | Enable rate limit middleware                                         | boolean  | false         |
| gin.middleware.rateLimit.ignore          | The paths of prefix that will be ignored by middleware               | []string | []            |
| gin.middleware.rateLimit.algorithm       | Provide algorithm, tokenBucket and leakyBucket are available options | string   | tokenBucket   |
| gin.middleware.rateLimit.reqPerSec       | Request per second globally                                          | int      | 0             |
| gin.middleware.rateLimit.paths.path      | Full path                                                            | string   | ""            |
| gin.middleware.rateLimit.paths.reqPerSec | Request per second by full path                                      | int      | 0             |

#### Timeout
| name                                   | description                                            | type     | default value |
|----------------------------------------|--------------------------------------------------------|----------|---------------|
| gin.middleware.timeout.enabled         | Enable timeout middleware                              | boolean  | false         |
| gin.middleware.timeout.ignore          | The paths of prefix that will be ignored by middleware | []string | []            |
| gin.middleware.timeout.timeoutMs       | Global timeout in milliseconds.                        | int      | 5000          |
| gin.middleware.timeout.paths.path      | Full path                                              | string   | ""            |
| gin.middleware.timeout.paths.timeoutMs | Timeout in milliseconds by full path                   | int      | 5000          |

#### Gzip
| name                        | description                                                                                                           | type     | default value      |
|-----------------------------|-----------------------------------------------------------------------------------------------------------------------|----------|--------------------|
| gin.middleware.gzip.enabled | Enable gzip middleware                                                                                                | boolean  | false              |
| gin.middleware.gzip.ignore  | The paths of prefix that will be ignored by middleware                                                                | []string | []                 |
| gin.middleware.gzip.level   | Provide level of compression, options are noCompression, bestSpeed, bestCompression, defaultCompression, huffmanOnly. | string   | defaultCompression |

#### CORS
| name                                 | description                                                            | type     | default value        |
|--------------------------------------|------------------------------------------------------------------------|----------|----------------------|
| gin.middleware.cors.enabled          | Enable cors middleware                                                 | boolean  | false                |
| gin.middleware.cors.ignore           | The paths of prefix that will be ignored by middleware                 | []string | []                   |
| gin.middleware.cors.allowOrigins     | Provide allowed origins with wildcard enabled.                         | []string | *                    |
| gin.middleware.cors.allowMethods     | Provide allowed methods returns as response header of OPTIONS request. | []string | All http methods     |
| gin.middleware.cors.allowHeaders     | Provide allowed headers returns as response header of OPTIONS request. | []string | Headers from request |
| gin.middleware.cors.allowCredentials | Returns as response header of OPTIONS request.                         | bool     | false                |
| gin.middleware.cors.exposeHeaders    | Provide exposed headers returns as response header of OPTIONS request. | []string | ""                   |
| gin.middleware.cors.maxAge           | Provide max age returns as response header of OPTIONS request.         | int      | 0                    |

#### JWT
> rk-gin using github.com/golang-jwt/jwt/v4, please beware of version compatibility.

In order to make swagger UI and RK tv work under JWT without JWT token, we need to ignore prefixes of paths as bellow.

```yaml
jwt:
  ...
  ignore:
   - "/sw"
```

| name                           | description                                                 | type     | default value          |
|--------------------------------|-------------------------------------------------------------|----------|------------------------|
| gin.middleware.jwt.enabled     | Enable JWT middleware                                       | boolean  | false                  |
| gin.middleware.jwt.ignore      | Provide ignoring path prefix.                               | []string | []                     |
| gin.middleware.jwt.signingKey  | Required, Provide signing key.                              | string   | ""                     |
| gin.middleware.jwt.signingKeys | Provide signing keys as scheme of <key>:<value>.            | []string | []                     |
| gin.middleware.jwt.signingAlgo | Provide signing algorithm.                                  | string   | HS256                  |
| gin.middleware.jwt.tokenLookup | Provide token lookup scheme, please see bellow description. | string   | "header:Authorization" |
| gin.middleware.jwt.authScheme  | Provide auth scheme.                                        | string   | Bearer                 |

The supported scheme of **tokenLookup** 

```
// Optional. Default value "header:Authorization".
// Possible values:
// - "header:<name>"
// - "query:<name>"
// - "param:<name>"
// - "cookie:<name>"
// - "form:<name>"
// Multiply sources example:
// - "header: Authorization,cookie: myowncookie"
```

#### Secure
| name                                        | description                                       | type     | default value   |
|---------------------------------------------|---------------------------------------------------|----------|-----------------|
| gin.middleware.secure.enabled               | Enable secure middleware                          | boolean  | false           |
| gin.middleware.secure.ignore                | Ignoring path prefix.                             | []string | []              |
| gin.middleware.secure.xssProtection         | X-XSS-Protection header value.                    | string   | "1; mode=block" |
| gin.middleware.secure.contentTypeNosniff    | X-Content-Type-Options header value.              | string   | nosniff         |
| gin.middleware.secure.xFrameOptions         | X-Frame-Options header value.                     | string   | SAMEORIGIN      |
| gin.middleware.secure.hstsMaxAge            | Strict-Transport-Security header value.           | int      | 0               |
| gin.middleware.secure.hstsExcludeSubdomains | Excluding subdomains of HSTS.                     | bool     | false           |
| gin.middleware.secure.hstsPreloadEnabled    | Enabling HSTS preload.                            | bool     | false           |
| gin.middleware.secure.contentSecurityPolicy | Content-Security-Policy header value.             | string   | ""              |
| gin.middleware.secure.cspReportOnly         | Content-Security-Policy-Report-Only header value. | bool     | false           |
| gin.middleware.secure.referrerPolicy        | Referrer-Policy header value.                     | string   | ""              |

#### CSRF
| name                               | description                                                                     | type     | default value         |
|------------------------------------|---------------------------------------------------------------------------------|----------|-----------------------|
| gin.middleware.csrf.enabled        | Enable csrf middleware                                                          | boolean  | false                 |
| gin.middleware.csrf.ignore         | Ignoring path prefix.                                                           | []string | []                    |
| gin.middleware.csrf.tokenLength    | Provide the length of the generated token.                                      | int      | 32                    |
| gin.middleware.csrf.tokenLookup    | Provide csrf token lookup rules, please see code comments for details.          | string   | "header:X-CSRF-Token" |
| gin.middleware.csrf.cookieName     | Provide name of the CSRF cookie. This cookie will store CSRF token.             | string   | _csrf                 |
| gin.middleware.csrf.cookieDomain   | Domain of the CSRF cookie.                                                      | string   | ""                    |
| gin.middleware.csrf.cookiePath     | Path of the CSRF cookie.                                                        | string   | ""                    |
| gin.middleware.csrf.cookieMaxAge   | Provide max age (in seconds) of the CSRF cookie.                                | int      | 86400                 |
| gin.middleware.csrf.cookieHttpOnly | Indicates if CSRF cookie is HTTP only.                                          | bool     | false                 |
| gin.middleware.csrf.cookieSameSite | Indicates SameSite mode of the CSRF cookie. Options: lax, strict, none, default | string   | default               |

### Full YAML
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
#    locale: "*::*::*::*"                                  # Optional, default: "*::*::*::*"
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
#    locale: "*::*::*::*"                                  # Optional, default: "*::*::*::*"
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
#    locale: "*::*::*::*"                                  # Optional, default: *::*::*::*
#    caPath: "certs/ca.pem"                                # Optional, default: ""
#    CertPemPath: "certs/server-cert.pem"                  # Optional, default: ""
#    KeyPemPath: "certs/server-key.pem"                    # Optional, default: ""
#config:
#  - name: my-config                                       # Required
#    description: "Description of entry"                   # Optional, default: ""
#    locale: "*::*::*::*"                                  # Optional, default: *::*::*::*
##    path: "config/config.yaml"                            # Optional
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
#        signingKey: "my-secret"                           # Required
#        ignore: [""]                                      # Optional, default: []
#        signingKeys:                                      # Optional
#          - "key:value"
#        signingAlgo: ""                                   # Optional, default: "HS256"
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

## Notice of V2
Master branch of this package is under upgrade which will be released to v2.x.x soon.

Major changes listed bellow. This will be updated with every commit.

| Last version | New version | Changes                                                                                                            |
|--------------|-------------|--------------------------------------------------------------------------------------------------------------------|
| v1.2.22      | v2          | TV is not supported because of LICENSE issue, new TV web UI will be released soon                                  |
| v1.2.22      | v2          | Remote repositry of ConfigEntry and CertEntry removed                                                              |
| v1.2.22      | v2          | Swagger json file and boot.yaml file could be embed into embed.FS and pass to rkentry                              |
| v1.2.22      | v2          | ZapLoggerEntry -> LoggerEntry                                                                                      |
| v1.2.22      | v2          | EventLoggerEntry -> EventEntry                                                                                     |
| v1.2.22      | v2          | LoggerEntry can be used as zap.Logger since all functions are inherited                                            |
| v1.2.22      | v2          | PromEntry can be used as prometheus.Registry since all functions are inherited                                     |
| v1.2.22      | v2          | rk-common dependency was removed                                                                                   |
| v1.2.22      | v2          | Entries are organized by EntryType instead of EntryName, so user can have same entry name with different EntryType |
| v1.2.22      | v2          | gin.middlewares -> gin.middleware in boot.yaml                                                                     |
| v1.2.22      | v2          | gin.middlewares.loggingZap -> gin.middleware.logging in boot.yaml                                                  |
| v1.2.22      | v2          | gin.middlewares.metricsProm -> gin.middleware.prom in boot.yaml                                                    |
| v1.2.22      | v2          | gin.middlewares.tracingTelemetry -> gin.middleware.trace in boot.yaml                                              |
| v1.2.22      | v2          | All middlewares are now support gin.middleware.xxx.ignorePrefix options in boot.yaml                               |
| v1.2.22      | v2          | Middlewares support gin.middleware.ignorePrefix in boot.yaml as global scope                                       |
| v1.2.22      | v2          | LoggerEntry, EventEntry, ConfigEntry, CertEntry now support locale to distinguish in differerent environment       |
| v1.2.22      | v2          | LoggerEntry, EventEntry, CertEntry can be referenced to gin entry in boot.yaml                                     |
| v1.2.22      | v2          | Healthy API was replaced by Ready and Alive which also provides validation func from user                          |
| v1.2.22      | v2          | DocsEntry was added into rk-entry                                                                                  |
| v1.2.22      | v2          | rk-entry support utility functions of embed.FS                                                                     |
| v1.2.22      | v2          | rk-entry bumped up to v2                                                                                           |

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


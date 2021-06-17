# rk-gin
Interceptor & bootstrapper designed for gin framework.
Currently, supports bellow interceptors

- auth
- logging
- metrics
- panic
- extension
- tracing

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Start Gin server from YAML config](#start-gin-server-from-yaml-config)
  - [Start Gin server from code](#start-gin-server-from-code)
  - [Logging Interceptor](#logging-interceptor)
  - [Metrics interceptor](#metrics-interceptor)
  - [Panic interceptor](#panic-interceptor)
  - [Auth interceptor](#auth-interceptor)
  - [Extension interceptor](#extension-interceptor)
  - [Tracing interceptor](#tracing-interceptor)
  - [Common Service](#common-service)
  - [TV Service](#tv-service)
  - [Development Status: Stable](#development-status-stable)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-gin`

## Quick Start
Bootstrapper can be used with YAML config.

### Start Gin server from YAML config
User can access common service with localhost:8080/sw
```yaml
---
rk: # NOT required
  appName: rk-example-entry           # Optional, default: "rkApp"
gin:
  - name: greeter                     # Required
    port: 8080                        # Required
    sw:                               # Optional
      enabled: true                   # Optional, default: false
    commonService:                    # Optional
      enabled: true                   # Optional, default: false
    interceptors:                     
      loggingZap:
        enabled: true                 # Optional, default: false
```

```go
func bootFromConfig() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/boot.yaml")

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntriesWithConfig("example/boot/boot.yaml")

	// Bootstrap gin entry
	go res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	res["greeter"].Interrupt(context.Background())
}
```

Available configuration
User can start multiple servers at the same time

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.name | Name of gin server entry | string | N/A |
| gin.port | Port of server | integer | nil, server won't start |
| gin.description | Description of server | string | "" |
| gin.cert.ref | Reference of cert entry declared in cert section | string | "" |
| gin.sw.enabled | Enable swagger | boolean | false | 
| gin.sw.path | Swagger path | string | / |
| gin.sw.jsonPath | Swagger json file path | string | / |
| gin.sw.headers | Headers will send with swagger response | array | [] |
| gin.commonService.enabled | Enable common service | boolean | false |
| gin.tv.enabled | Enable RK TV whose path is /rk/v1/tv | boolean | false |
| gin.prom.enabled | Enable prometheus | boolean | false |
| gin.prom.path | Path of prometheus | string | metrics |
| gin.prom.cert.ref |  Reference of cert entry declared in cert section | string | "" |
| gin.prom.pusher.enabled | Enable prometheus pusher | bool | false |
| gin.prom.pusher.jobName | Job name would be attached as label while pushing to remote pushgateway | string | "" |
| gin.prom.pusher.remoteAddress | PushGateWay address, could be form of http://x.x.x.x or x.x.x.x | string | "" |
| gin.prom.pusher.intervalMs | Push interval in milliseconds | string | 1000 |
| gin.prom.pusher.basicAuth | Basic auth used to interact with remote pushgateway, form of \<user:pass\> | string | "" |
| gin.prom.pusher.cert.ref | Reference of rkentry.CertEntry | string | "" |
| gin.logger.zapLogger.ref | Reference of logger entry declared above | string | "" |
| gin.logger.eventLogger.ref | Reference of logger entry declared above | string | "" |
| gin.interceptors.loggingZap.enabled | Enable logging interceptor | boolean | false |
| gin.interceptors.metricsProm.enabled | Enable prometheus metrics for every request | boolean | false |
| gin.interceptors.basicAuth.enabled | Enable auth interceptor | boolean | false |
| gin.interceptors.basicAuth.credentials | Provide basic auth credentials, form of \<user:pass\> | string | false |
| gin.interceptors.extension.enabled | Enable extension interceptor | boolean | false |
| gin.interceptors.extension.prefix | Prefix of extension header key | string | rk |
| gin.interceptors.tracingTelemetry.enabled | Enable tracing interceptor with opentelemetry | bool | false |
| gin.interceptors.tracingTelemetry.exporter.file.enabled | Enable exporter which will write tracing info to file or stdout | string | stdout |
| gin.interceptors.tracingTelemetry.exporter.file.outputPath | Output path of tracing log | string | stdout |
| gin.interceptors.tracingTelemetry.exporter.jaeger.enabled | Enable exporter which will write tracing info to jaeger agent | bool | false |
| gin.interceptors.tracingTelemetry.exporter.jaeger.agentEndpoint | Jaeger agent endpoint | string | "localhost:6832" |

Interceptors can be used with chain.

### Start Gin server from code

```go
func bootFromCode() {
	// Create gin entry
	entry := rkgin.RegisterGinEntry(
		rkgin.WithNameGin("greeter"),
		rkgin.WithPortGin(8080),
		rkgin.WithCommonServiceEntryGin(rkgin.NewCommonServiceEntry()),
		rkgin.WithInterceptorsGin(rkginlog.LoggingZapInterceptor([]rkginlog.Option{}...)))

	// Start server
	go entry.Bootstrap(context.Background())

	// Wait for shutdown sig
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt server
	entry.Interrupt(context.Background())
}
```

### Logging Interceptor
Logging interceptor uses [zap logger](https://github.com/uber-go/zap) and [rk-query](https://github.com/rookie-ninja/rk-query) logs every request.
[rk-prom](https://github.com/rookie-ninja/rk-prom) also used for prometheus metrics.

```go
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginbasic.BasicInterceptor(),
		rkginlog.LoggingZapInterceptor(
			rkginlog.WithEventFactory(rkquery.NewEventFactory()),
			rkginlog.WithLogger(rklogger.StdoutLogger)),
		rkginpanic.PanicInterceptor(),
	)

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
```

Output: 
```log
------------------------------------------------------------------------
endTime=2021-06-15T02:46:00.256757+08:00
startTime=2021-06-15T02:46:00.256704+08:00
elapsedNano=53081
timezone=CST
ids={"eventId":"ab6695d3-e698-4434-8121-c0c21e4451b4"}
app={"appName":"unknown","appVersion":"unknown","entryName":"rkEntry","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/hello","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:61435
operation=/hello
resCode=200
eventStatus=Ended
EOE
```

### Metrics interceptor
```go
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginbasic.BasicInterceptor(),
        rkginmetrics.MetricsPromInterceptor(),
		rkginpanic.PanicInterceptor(),
	)

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
```


### Panic interceptor
```go
func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginlog.LoggingZapInterceptor(
			rkginlog.WithEventFactory(rkquery.NewEventFactory()),
			rkginlog.WithLogger(rklogger.StdoutLogger)),
		rkginpanic.PanicInterceptor())

	router.GET("/hello", func(ctx *gin.Context) {
		panic(errors.New(""))
	})
	router.Run(":8080")
}
```
Output
```log
------------------------------------------------------------------------
endTime=2021-06-15T02:47:10.368031+08:00
startTime=2021-06-15T02:47:10.367417+08:00
elapsedNano=614458
timezone=CST
ids={"eventId":"310a4255-1de4-4607-9eed-67c0e858864f"}
app={"appName":"unknown","appVersion":"unknown","entryName":"rkEntry","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/hello","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:61440
operation=/hello
resCode=500
eventStatus=Ended
EOE
```

### Auth interceptor
```go
func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginlog.LoggingZapInterceptor(
			rkginlog.WithEventFactory(rkquery.NewEventFactory()),
			rkginlog.WithLogger(rklogger.StdoutLogger)),
		rkginauth.BasicAuthInterceptor(gin.Accounts{"user": "pass"}, "realm"),
		rkginpanic.PanicInterceptor())

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}
```
Output
```log
------------------------------------------------------------------------
endTime=2021-06-15T02:46:00.256757+08:00
startTime=2021-06-15T02:46:00.256704+08:00
elapsedNano=53081
timezone=CST
ids={"eventId":"ab6695d3-e698-4434-8121-c0c21e4451b4"}
app={"appName":"unknown","appVersion":"unknown","entryName":"rkEntry","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/hello","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:61435
operation=/hello
resCode=200
eventStatus=Ended
EOE
```

### Extension interceptor
This interceptor will add bellow headers as extension in response.

- X-[Prefix]-Request-Id: Request id generated by the interceptor.
- X-[Prefix]-Location: A valid URI.
- X-[Prefix]-Locale: Locale of current service.
- X-[Prefix]-App: Application name.
- X-[Prefix]-App-Version: Version of application.
- X-[Prefix]-App-Unix-Time: Unix time of current application.
- X-[Prefix]-Request-Received-Time: Time of current request received by application.

```shell script
$ curl -vs -X GET "http://localhost:8080/rk/v1/configs" -H  "accept: application/json"
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /rk/v1/configs HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.64.1
> accept: application/json
>
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Rk-App-Name: rkApp
< X-Rk-App-Unix-Time: 2021-05-31T01:38:51.368372+08:00
< X-Rk-App-Version: v0.0.0
< X-Rk-Locale: *::*::*::*
< X-Rk-Location: http://localhost:8080/rk/v1/configs
< X-Rk-Request-Id: 75b982e4-c841-4dc4-9994-e14152530311
< X-Rk-Request-Received-Time: 2021-05-31T01:38:51.368372+08:00
< Date: Sun, 30 May 2021 17:38:51 GMT
< Content-Length: 14
<
* Connection #0 to host localhost left intact
{"entries":[]}
```

### Tracing interceptor
This interceptor will automatically collect tracing spans and export to specified exporter.
- File exporter (Write trace log to files or stdout)
- Jaeger exporter (Write to jaeger agent)

```go
func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginbasic.BasicInterceptor(),
		rkgintrace.TelemetryInterceptor(),
		rkginpanic.PanicInterceptor(),
	)

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}
```

```json
[
        {
                "SpanContext": {
                        "TraceID": "b8af9be0722842783499a2b4756af62c",
                        "SpanID": "0bf79b2c5e49bcb6",
                        "TraceFlags": "01",
                        "TraceState": null,
                        "Remote": false
                },
                "Parent": {
                        "TraceID": "00000000000000000000000000000000",
                        "SpanID": "0000000000000000",
                        "TraceFlags": "00",
                        "TraceState": null,
                        "Remote": false
                },
                "SpanKind": 2,
                "Name": "/hello",
                "StartTime": "2021-06-15T02:51:25.349015+08:00",
                "EndTime": "2021-06-15T02:51:25.349068874+08:00",
                "Attributes": [
                        {
                                "Key": "net.transport",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "IP.TCP"
                                }
                        },
                        {
                                "Key": "net.peer.name",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "[::1]"
                                }
                        },
                        {
                                "Key": "net.peer.port",
                                "Value": {
                                        "Type": "INT64",
                                        "Value": 61452
                                }
                        },
                        {
                                "Key": "net.host.name",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "localhost"
                                }
                        },
                        {
                                "Key": "net.host.port",
                                "Value": {
                                        "Type": "INT64",
                                        "Value": 8080
                                }
                        },
                        {
                                "Key": "http.method",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "GET"
                                }
                        },
                        {
                                "Key": "http.target",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "/hello"
                                }
                        },
                        {
                                "Key": "http.server_name",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "rkApp"
                                }
                        },
                        {
                                "Key": "http.route",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "/hello"
                                }
                        },
                        {
                                "Key": "http.user_agent",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "curl/7.64.1"
                                }
                        },
                        {
                                "Key": "http.scheme",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "http"
                                }
                        },
                        {
                                "Key": "http.host",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "localhost:8080"
                                }
                        },
                        {
                                "Key": "http.flavor",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "1.1"
                                }
                        },
                        {
                                "Key": "http.status_code",
                                "Value": {
                                        "Type": "INT64",
                                        "Value": 200
                                }
                        }
                ],
                "MessageEvents": null,
                "Links": null,
                "StatusCode": "Unset",
                "StatusMessage": "",
                "DroppedAttributeCount": 0,
                "DroppedMessageEventCount": 0,
                "DroppedLinkCount": 0,
                "ChildSpanCount": 0,
                "Resource": [
                        {
                                "Key": "service.entryName",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "rkEntry"
                                }
                        },
                        {
                                "Key": "service.entryType",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "gin"
                                }
                        },
                        {
                                "Key": "service.name",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "rkApp"
                                }
                        },
                        {
                                "Key": "service.version",
                                "Value": {
                                        "Type": "STRING",
                                        "Value": "v0.0.0"
                                }
                        }
                ],
                "InstrumentationLibrary": {
                        "Name": "rkEntry",
                        "Version": "semver:0.20.0"
                }
        }
]
```

### Common Service

| path | description |
| ------ | ------ |
| /rk/v1/apis | List API |
| /rk/v1/certs | List CertEntry |
| /rk/v1/configs | List ConfigEntry |
| /rk/v1/deps | List dependencies related application |
| /rk/v1/entries | List all Entry |
| /rk/v1/gc | Trigger GC |
| /rk/v1/healthy | Get application healthy status, returns true if application is running |
| /rk/v1/info | Get application and process info |
| /rk/v1/license | Get license related application |
| /rk/v1/logs | List logger related entries |
| /rk/v1/readme | Get README file |
| /rk/v1/req | List prometheus metrics of requests |
| /rk/v1/sys | Get OS stat |
| /rk/v1/tv | Get HTML page of /tv |

### TV Service

| path | description |
| ------ | ------ |
| /rk/v1/tv or /rk/v1/tv/overview | Get application and process info of HTML page |
| /rk/v1/tv/api | Get API of HTML page |
| /rk/v1/tv/entry | Get entry of HTML page |
| /rk/v1/tv/config | Get config of HTML page |
| /rk/v1/tv/cert | Get cert of HTML page |
| /rk/v1/tv/os | Get OS of HTML page |
| /rk/v1/tv/env | Get Go environment of HTML page |
| /rk/v1/tv/prometheus | Get metrics of HTML page |
| /rk/v1/log | Get log of HTML page |
| /rk/v1/dep | Get dependency of HTML page |

### Development Status: Stable

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
dongxuny@gmail.com. That email list is a private, safe space; even the zap
maintainers don't have access, so don't hesitate to hold us to a high
standard.

Released under the [MIT License](LICENSE).


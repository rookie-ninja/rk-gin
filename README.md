# rk-gin
Interceptor & bootstrapper designed for gin framework.
Currently, supports bellow interceptors

- logging & metrics
- auth
- panic
- bootstrapper

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Start Gin server from YAML config](#start-gin-server-from-yaml-config)
  - [Start Gin server from code](#start-gin-server-from-code)
  - [Logging & Metrics interceptor](#logging--metrics-interceptor)
  - [Panic interceptor](#panic-interceptor)
  - [Auth interceptor](#auth-interceptor)
  - [Common Services](#common-services)
  - [Development Status: Stable](#development-status-stable)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-gin`

## Quick Start
Bootstrapper can be used with YAML config

### Start Gin server from YAML config
User can access common service with localhost:8080/sw
```yaml
---
gin:
  - name: greeter                                    # Required
    port: 8080                                       # Required
    cert:
      ref: "local-test"                              # Optional, default: "", reference of cert entry declared above
    sw:
      enabled: true                                  # Optional, default: false
      path: "sw"                                     # Optional, default: "sw"
      headers: ["sw:rk"]                             # Optional, default: []
    commonService:
      enabled: true                                  # Optional, default: false
      pathPrefix: "/v1/rk/"                          # Optional, default: "/v1/rk/"
    tv:
      enabled:  true                                 # Optional, default: false
      pathPrefix: "/v1/rk/"                          # Optional, default: "/v1/rk/"
    prom:
      enabled: true                                  # Optional, default: false
      path: "metrics"                                # Optional, default: ""
      cert:
        ref: "local-test"                            # Optional, default: "", reference of cert entry declared above
      pusher:
        enabled: false                               # Optional, default: false
        jobName: "greeter-pusher"                    # Required
        remoteAddress: "localhost:9091"              # Required
        basicAuth: "user:pass"                       # Optional, default: ""
        intervalMS: 1000                             # Optional, default: 1000
        cert:
          ref: "local-test"                          # Optional, default: "", reference of cert entry declared above
    logger:
      zapLogger:
        ref: zap-logger                              # Optional, default: logger of STDOUT, reference of logger entry declared above
      eventLogger:
        ref: event-logger                            # Optional, default: logger of STDOUT, reference of logger entry declared above
    interceptors:
      loggingZap:
        enabled: true                                # Optional, default: false
      metricsProm:
        enabled: true                                # Optional, default: false
      basicAuth:
        enabled: true                                # Optional, default: false
        credentials:
          - "user:pass"                              # Optional, default: ""
```

```go
func bootFromConfig() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterBasicEntriesFromConfig("example/boot/boot.yaml")

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
| gin.cert.ref | Reference of cert entry declared in cert section | string | "" |
| gin.sw.enabled | Enable swagger | boolean | false | 
| gin.sw.path | Swagger path | string | / |
| gin.sw.jsonPath | Swagger json file path | string | / |
| gin.sw.headers | Headers will send with swagger response | array | [] |
| gin.commonService.enable | Enable common service | boolean | false |
| gin.commonService.pathPrefix | Path prefix of common service | string | /v1/rk |
| gin.tv.enable | Enable RK TV whose path is /v1/rk/tv | boolean | false |
| gin.tv.pathPrefix | Path prefix of common service | string | /v1/rk |
| gin.prom.enable | Enable prometheus | boolean | false |
| gin.prom.path | Path of prometheus | string | metrics |
| gin.prom.cert.ref |  Reference of cert entry declared in cert section | string | "" |
| gin.prom.path | Path of prometheus | string | metrics |
| gin.prom.pusher.enabled | Enable prometheus pusher | bool | false |
| gin.prom.pusher.jobName | Job name would be attached as label while pushing to remote pushgateway | string | "" |
| gin.prom.pusher.remoteAddress | Pushgateway address, could be form of http://x.x.x.x or x.x.x.x | string | "" |
| gin.prom.pusher.basicAuth | Basic auth used to interact with remote pushgateway, form of \<user:pass\> | string | "" |
| gin.prom.pusher.cert.ref | Reference of rkentry.CertEntry | string | "" |
| gin.prom.cert.ref | Reference of rkentry.CertEntry | string | "" |
| gin.logger.zapLogger.ref | Reference of logger entry declared above | string | "" |
| gin.logger.eventLogger.ref | Reference of logger entry declared above | string | "" |
| gin.interceptors.loggingZap.enabled | Enable logging interceptor | boolean | false |
| gin.interceptors.metricsProm.enabled | Enable prometheus metrics for every request | boolean | false |
| gin.interceptors.basicAuth.enabled | Enable auth interceptor | boolean | false |
| gin.interceptors.basicAuth.credentials | Provide basic auth credentials, form of \<user:pass\> | string | false |

Interceptors can be used with chain.

### Start Gin server from code

```go
func bootFromCode() {
	// Create event data
	fac := rkquery.NewEventFactory()

	// Create options for interceptor
	opts := []rkginlog.Option{
		rkginlog.WithEventFactory(fac),
		rkginlog.WithLogger(rklogger.StdoutLogger),
	}

	// Create gin entry
	entry := rkgin.RegisterGinEntry(
		rkgin.WithNameGin("greeter"),
		rkgin.WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		rkgin.WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()),
		rkgin.WithPortGin(8080),
		rkgin.WithCommonServiceEntryGin(rkgin.NewCommonServiceEntry()),
		rkgin.WithTVEntryGin(rkgin.NewTVEntry()),
		rkgin.WithInterceptorsGin(rkginlog.LoggingZapInterceptor(opts...)))

	// Start server
	go entry.Bootstrap(context.Background())

	// Wait for shutdown sig
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt server
	entry.Interrupt(context.Background())
}
```

### Logging & Metrics interceptor
Logging interceptor uses [zap logger](https://github.com/uber-go/zap) and [rk-query](https://github.com/rookie-ninja/rk-query) logs every requests.
[rk-prom](https://github.com/rookie-ninja/rk-prom) also used for prometheus metrics.

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
		//ctx.String(http.StatusOK, "Hello world")
		panic(errors.New(""))
	})
	router.Run(":8080")
}

```

Output: 
```log
------------------------------------------------------------------------
end_time=2020-11-06T01:31:36.372368+08:00
start_time=2020-11-06T01:31:36.372265+08:00
time=0
hostname=JEREMYYIN-MB0
timing={}
counter={}
pair={}
error={}
field={"api.method":"GET","api.path":"/hello","api.protocol":"HTTP/1.1","api.query":"","app_version":"latest","az":"unknown","domain":"unknown","elapsed_ms":0,"end_time":"2020-11-06T01:31:36.372368+08:00","incoming_request_ids":[],"local.IP":"10.8.0.2","outgoing_request_id":[],"realm":"unknown","region":"unknown","remote.IP":"localhost","remote.port":"61210","res_code":200,"start_time":"2020-11-06T01:31:36.372265+08:00","user_agent":"curl/7.49.1"}
remote_addr=localhost:61210
app_name=Unknown
operation=GET-/hello
event_status=Ended
res_code=200
timezone=CST
os=darwin
arch=amd64
EOE
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
		rkginauth.BasicAuthInterceptor(gin.Accounts{"user": "pass"}, "realm"),
		rkginpanic.PanicInterceptor())

	router.GET("/hello", func(ctx *gin.Context) {
		//ctx.String(http.StatusOK, "Hello world")
		panic(errors.New(""))
	})
	router.Run(":8080")
}
```
Output
```log
------------------------------------------------------------------------
end_time=2020-11-02T04:16:10.927366+08:00
start_time=2020-11-02T04:16:10.927095+08:00
time=0
hostname=JEREMYYIN-MB0
timing={}
counter={}
pair={}
error={}
field={"api.method":"GET","api.path":"/hello","api.protocol":"HTTP/1.1","api.query":"","app_version":"latest","az":"unknown","domain":"unknown","elapsed_ms":0,"end_time":"2020-11-02T04:16:10.927372+08:00","incoming_request_ids":[],"local.IP":"192.168.3.26","outgoing_request_id":[],"realm":"unknown","region":"unknown","remote.IP":"localhost","remote.port":"56567","request":"GET /hello HTTP/1.1\r\nHost: localhost:8080\r\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9\r\nAccept-Encoding: gzip, deflate, br\r\nAccept-Language: en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,ko;q=0.6,zh-TW;q=0.5,ja;q=0.4,vi;q=0.3,es;q=0.2\r\nAuthorization: Basic dXNlcjpwYXNz\r\nCache-Control: max-age=0\r\nConnection: keep-alive\r\nCookie: Goland-b0e6b6d4=d7c5eb18-1c4b-446e-8a61-bd60e69342bc\r\nSec-Fetch-Dest: document\r\nSec-Fetch-Mode: navigate\r\nSec-Fetch-Site: none\r\nSec-Fetch-User: ?1\r\nUpgrade-Insecure-Requests: 1\r\nUser-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36\r\n\r\n","res_code":500,"stack":"goroutine 40 [running]:\nruntime/debug.Stack(0x0, 0xc0003a0000, 0x2de)\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x9d\ngithub.com/rookie-ninja/rk-gin-interceptor/panic/zap.RkGinPanicZap.func1.1(0x1899220, 0xc00039c0f0, 0xc00039c000)\n\t/Users/donghun221/workspace/rk/rk-gin-interceptor/panic/zap/interceptor.go:58 +0x6f8\npanic(0x16a4e60, 0xc0003674d0)\n\t/usr/local/go/src/runtime/panic.go:969 +0x166\nmain.main.func1(0xc00039c000)\n\t/Users/donghun221/workspace/rk/rk-gin-interceptor/example/main.go:25 +0x59\ngithub.com/gin-gonic/gin.(*Context).Next(0xc00039c000)\n\t/Users/donghun221/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/context.go:161 +0x3b\ngithub.com/rookie-ninja/rk-gin-interceptor/panic/zap.RkGinPanicZap.func1(0xc00039c000)\n\t/Users/donghun221/workspace/rk/rk-gin-interceptor/panic/zap/interceptor.go:66 +0x79\ngithub.com/gin-gonic/gin.(*Context).Next(0xc00039c000)\n\t/Users/donghun221/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/context.go:161 +0x3b\ngithub.com/rookie-ninja/rk-gin-interceptor/logging/zap.RkGinZap.func1(0xc00039c000)\n\t/Users/donghun221/workspace/rk/rk-gin-interceptor/logging/zap/interceptor.go:46 +0xd90\ngithub.com/gin-gonic/gin.(*Context).Next(0xc00039c000)\n\t/Users/donghun221/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/context.go:161 +0x3b\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest(0xc00014c000, 0xc00039c000)\n\t/Users/donghun221/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/gin.go:409 +0x666\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP(0xc00014c000, 0x1881880, 0xc00039a0e0, 0xc00022ad00)\n\t/Users/donghun221/go/pkg/mod/github.com/gin-gonic/gin@v1.6.3/gin.go:367 +0x14d\nnet/http.serverHandler.ServeHTTP(0xc00014a0e0, 0x1881880, 0xc00039a0e0, 0xc00022ad00)\n\t/usr/local/go/src/net/http/server.go:2836 +0xa3\nnet/http.(*conn).serve(0xc00013a0a0, 0x1883800, 0xc00038b440)\n\t/usr/local/go/src/net/http/server.go:1924 +0x86c\ncreated by net/http.(*Server).Serve\n\t/usr/local/go/src/net/http/server.go:2962 +0x35c\n","start_time":"2020-11-02T04:16:10.927095+08:00","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"}
remote_addr=localhost:56567
app_name=Unknown
operation=GET-/hello
event_status=Ended
res_code=500
timezone=CST
os=darwin
arch=amd64
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
		//ctx.String(http.StatusOK, "Hello world")
		panic(errors.New(""))
	})
	router.Run(":8080")
}
```
Output
```log
------------------------------------------------------------------------
end_time=2020-11-06T01:34:05.541346+08:00
start_time=2020-11-06T01:34:05.54126+08:00
time=0
hostname=JEREMYYIN-MB0
timing={}
counter={}
pair={}
error={}
field={"api.method":"GET","api.path":"/hello","api.protocol":"HTTP/1.1","api.query":"","app_version":"latest","az":"unknown","domain":"unknown","elapsed_ms":0,"end_time":"2020-11-06T01:34:05.541346+08:00","incoming_request_ids":[],"local.IP":"10.8.0.2","outgoing_request_id":[],"realm":"unknown","region":"unknown","remote.IP":"localhost","remote.port":"61231","res_code":401,"start_time":"2020-11-06T01:34:05.54126+08:00","user_agent":"curl/7.49.1"}
remote_addr=localhost:61231
app_name=Unknown
operation=GET-/hello
event_status=Ended
res_code=401
timezone=CST
os=darwin
arch=amd64
EOE
```

### Common Services
User can start multiple servers at the same time

| path | description |
| ------ | ------ |
| /v1/rk/healthy | Always return true if service is available |
| /v1/rk/gc | Trigger gc and return memory stats |
| /v1/rk/info | Return basic info |
| /v1/rk/config | Return configs in memory |
| /v1/rk/apis | List all apis |
| /v1/rk/sys | Return system information including cpu and memory usage |
| /v1/rk/req | Return requests stats recorded by prometheus client |
| /v1/rk/tv | Web ui for metrics |

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


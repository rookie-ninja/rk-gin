# rk-gin
[![build](https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-gin/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-gin/branch/master/graph/badge.svg?token=S0B7KTMIHW)](https://codecov.io/gh/rookie-ninja/rk-gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-gin)](https://goreportcard.com/report/github.com/rookie-ninja/rk-gin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Interceptor & bootstrapper designed for gin framework. Currently, supports bellow functionalities.

| Name | Description |
| ---- | ---- |
| Start with YAML | Start service with YAML config. |
| Start with code | Start service from code. |
| Gin Service | Gin service. |
| Swagger Service | Swagger UI. |
| Common Service | List of common API available on Gin. |
| TV Service | A Web UI shows application and environment information. |
| Metrics interceptor | Collect RPC metrics and export as prometheus client. |
| Log interceptor | Log every RPC requests as event with rk-query. |
| Trace interceptor | Collect RPC trace and export it to stdout, file or jaeger. |
| Panic interceptor | Recover from panic for RPC requests and log it. |
| Meta interceptor | Send application metadata as header to client. |
| Auth interceptor | Support [Basic Auth] and [API Key] authorization types. |

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Start Gin Service](#start-gin-service)
  - [Output](#output)
    - [Gin Service](#gin-service)
    - [Swagger Service](#swagger-service)
    - [Swagger Service](#swagger-service-1)
    - [TV Service](#tv-service)
    - [Metrics](#metrics)
    - [Logging](#logging)
    - [Meta](#meta)
- [YAML Config](#yaml-config)
  - [Gin Service](#gin-service-1)
  - [Common Service](#common-service)
  - [Swagger Service](#swagger-service-2)
  - [Prom Client](#prom-client)
  - [TV Service](#tv-service-1)
  - [Interceptors](#interceptors)
    - [Log](#log)
    - [Metrics](#metrics-1)
    - [Auth](#auth)
    - [Meta](#meta-1)
    - [Tracing](#tracing)
  - [Development Status: Stable](#development-status-stable)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-gin`

## Quick Start
Bootstrapper can be used with YAML config. In the bellow example, we will start bellow services automatically.
- Gin Service
- Swagger Service
- Common Service
- TV Service
- Metrics
- Logging
- Meta

Please refer example at [example/boot/simple](example/boot/simple).

### Start Gin Service

- [boot.yaml](example/boot/simple/boot.yaml)

```yaml
---
gin:
  - name: greeter                     # Required
    port: 8080                        # Required
    enabled: true                     # Required
    tv:
      enabled: true                   # Optional, default: false
    prom:
      enabled: true                   # Optional, default: false
    sw:                               # Optional
      enabled: true                   # Optional, default: false
    commonService:                    # Optional
      enabled: true                   # Optional, default: false
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      meta:
        enabled: true
```

- [main.go](example/boot/simple/main.go)

```go
func main() {
    // Bootstrap basic entries from boot config.
    rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

    // Bootstrap gin entry from boot config
    res := rkgin.RegisterGinEntriesWithConfig("example/boot/simple/boot.yaml")

    // Bootstrap gin entry
    res["greeter"].Bootstrap(context.Background())

    // Wait for shutdown signal
    rkentry.GlobalAppCtx.WaitForShutdownSig()

    // Interrupt gin entry
    res["greeter"].Interrupt(context.Background())
}
```

```go
$ go run main.go
```

### Output
#### Gin Service
Try to test Gin Service with [curl](https://curl.se/)
```shell script
# Curl to common service
$ curl localhost:8080/rk/v1/healthy
{"healthy":true}
```

#### Swagger Service
By default, we could access swagger UI at [/sw].
- http://localhost:8080/sw

![sw](docs/img/simple-sw.png)

#### Swagger Service
By default, we could access swagger UI at [/sw].
- http://localhost:8080/sw

![sw](docs/img/simple-sw.png)

#### TV Service
By default, we could access TV at [/tv].

![tv](docs/img/simple-tv.png)

#### Metrics
By default, we could access prometheus client at [/metrics]
- http://localhost:8080/metrics

![prom](docs/img/simple-prom.png)

#### Logging
By default, we enable zap logger and event logger with console encoding type.
```shell script
2021-06-25T01:22:23.907+0800    INFO    Bootstrapping SwEntry.  {"eventId": "0f056bf9-0811-4fdb-b1eb-8d01b3b2a576", "entryName": "greeter-sw", "entryType": "GinSwEntry", "jsonPath": "", "path": "/sw/", "port": 8080}
2021-06-25T01:22:23.907+0800    INFO    Bootstrapping promEntry.        {"eventId": "0f056bf9-0811-4fdb-b1eb-8d01b3b2a576", "entryName": "greeter-prom", "entryType": "GinPromEntry", "entryDescription": "Internal RK entry which implements prometheus client with Gin framework.", "path": "/metrics", "port": 8080}
2021-06-25T01:22:23.907+0800    INFO    Bootstrapping CommonServiceEntry.       {"eventId": "0f056bf9-0811-4fdb-b1eb-8d01b3b2a576", "entryName": "greeter-commonService", "entryType": "GinCommonServiceEntry"}
2021-06-25T01:22:23.909+0800    INFO    Bootstrapping tvEntry.  {"eventId": "0f056bf9-0811-4fdb-b1eb-8d01b3b2a576", "entryName": "greeter-tv", "entryType": "GinTvEntry", "path": "/rk/v1/tv/*item"}
2021-06-25T01:22:23.909+0800    INFO    Bootstrapping GinEntry. {"eventId": "0f056bf9-0811-4fdb-b1eb-8d01b3b2a576", "entryName": "greeter", "entryType": "GinEntry", "port": 8080, "interceptorsCount": 6, "swEnabled": true, "tlsEnabled": false, "commonServiceEnabled": true, "tvEnabled": true, "swPath": "/sw/", "promPath": "/metrics", "promPort": 8080}
```
```shell script
------------------------------------------------------------------------
endTime=2021-06-25T01:22:23.90783+08:00
startTime=2021-06-25T01:22:23.90781+08:00
elapsedNano=20378
timezone=CST
ids={"eventId":"0f056bf9-0811-4fdb-b1eb-8d01b3b2a576"}
app={"appName":"rk-gin","appVersion":"master-xxx","entryName":"greeter-sw","entryType":"GinSwEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"entryName":"greeter-sw","entryType":"GinSwEntry","jsonPath":"","path":"/sw/","port":8080}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
...
------------------------------------------------------------------------
endTime=2021-06-25T01:22:23.909337+08:00
startTime=2021-06-25T01:22:23.907776+08:00
elapsedNano=1560406
timezone=CST
ids={"eventId":"0f056bf9-0811-4fdb-b1eb-8d01b3b2a576"}
app={"appName":"rk-gin","appVersion":"master-xxx","entryName":"greeter","entryType":"GinEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"commonServiceEnabled":true,"entryName":"greeter","entryType":"GinEntry","interceptorsCount":6,"port":8080,"promPath":"/metrics","promPort":8080,"swEnabled":true,"swPath":"/sw/","tlsEnabled":false,"tvEnabled":true}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
```

#### Meta
By default, we will send back some metadata to client including gateway with headers.
```shell script
$ curl -vs localhost:8080/rk/v1/healthy
...
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Request-Id: 3332e575-43d8-4bfe-84dd-45b5fc5fb104
< X-Rk-App-Name: rk-gin
< X-Rk-App-Unix-Time: 2021-06-25T01:30:45.143869+08:00
< X-Rk-App-Version: master-xxx
< X-Rk-Received-Time: 2021-06-25T01:30:45.143869+08:00
< X-Trace-Id: 65b9aa7a9705268bba492fdf4a0e5652
< Date: Thu, 24 Jun 2021 17:30:45 GMT
...
```

## YAML Config
Available configuration
User can start multiple gin servers at the same time. Please make sure use different port and name.

### Gin Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.name | The name of gin server | string | N/A |
| gin.port | The port of gin server | integer | nil, server won't start |
| gin.enabled | Enable Gin entry or not | bool | false |
| gin.description | Description of gin entry. | string | "" |
| gin.cert.ref | Reference of cert entry declared in [cert entry](https://github.com/rookie-ninja/rk-entry#certentry) | string | "" |
| gin.logger.zapLogger.ref | Reference of zapLoggerEntry declared in [zapLoggerEntry](https://github.com/rookie-ninja/rk-entry#zaploggerentry) | string | "" |
| gin.logger.eventLogger.ref | Reference of eventLoggerEntry declared in [eventLoggerEntry](https://github.com/rookie-ninja/rk-entry#eventloggerentry) | string | "" |

### Common Service
| Path | Description |
| ---- | ---- |
| /rk/v1/apis | List APIs in current GinEntry. |
| /rk/v1/certs | List CertEntry. |
| /rk/v1/configs | List ConfigEntry. |
| /rk/v1/deps | List dependencies related application, entire contents of go.mod file would be returned. |
| /rk/v1/entries | List all Entries. |
| /rk/v1/gc | Trigger GC |
| /rk/v1/healthy | Get application healthy status. |
| /rk/v1/info | Get application and process info. |
| /rk/v1/license | Get license related application, entire contents of LICENSE file would be returned. |
| /rk/v1/logs | List logger related entries. |
| /rk/v1/git | Get git information. |
| /rk/v1/readme | Get contents of README file. |
| /rk/v1/req | List prometheus metrics of requests. |
| /rk/v1/sys | Get OS stat. |
| /rk/v1/tv | Get HTML page of /tv. |

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.commonService.enabled | Enable embedded common service | boolean | false |

### Swagger Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.sw.enabled | Enable swagger service over gin server | boolean | false |
| gin.sw.path | The path access swagger service from web | string | /sw |
| gin.sw.jsonPath | Where the swagger.json files are stored locally | string | "" |
| gin.sw.headers | Headers would be sent to caller as scheme of [key:value] | []string | [] |

### Prom Client
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.prom.enabled | Enable prometheus | boolean | false |
| gin.prom.path | Path of prometheus | string | /metrics |
| gin.prom.pusher.enabled | Enable prometheus pusher | bool | false |
| gin.prom.pusher.jobName | Job name would be attached as label while pushing to remote pushgateway | string | "" |
| gin.prom.pusher.remoteAddress | PushGateWay address, could be form of http://x.x.x.x or x.x.x.x | string | "" |
| gin.prom.pusher.intervalMs | Push interval in milliseconds | string | 1000 |
| gin.prom.pusher.basicAuth | Basic auth used to interact with remote pushgateway, form of [user:pass] | string | "" |
| gin.prom.pusher.cert.ref | Reference of rkentry.CertEntry | string | "" |

### TV Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.tv.enabled | Enable RK TV | boolean | false |

### Interceptors
#### Log
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.interceptors.loggingZap.enabled | Enable log interceptor | boolean | false |
| gin.interceptors.loggingZap.zapLoggerEncoding | json or console | string | console |
| gin.interceptors.loggingZap.zapLoggerOutputPaths | Output paths | []string | stdout |
| gin.interceptors.loggingZap.eventLoggerEncoding | json or console | string | console |
| gin.interceptors.loggingZap.eventLoggerOutputPaths | Output paths | []string | false |

We will log two types of log for every RPC call.
- zapLogger

Contains user printed logging with requestId or traceId.

- eventLogger

Contains per RPC metadata, response information, environment information and etc.

| Field | Description |
| ---- | ---- |
| endTime | As name described |
| startTime | As name described |
| elapsedNano | Elapsed time for RPC in nanoseconds |
| timezone | As name described |
| ids | Contains three different ids(eventId, requestId and traceId). If meta interceptor was enabled or event.SetRequestId() was called by user, then requestId would be attached. eventId would be the same as requestId if meta interceptor was enabled. If trace interceptor was enabled, then traceId would be attached. |
| app | Contains [appName, appVersion](https://github.com/rookie-ninja/rk-entry#appinfoentry), entryName, entryType. |
| env | Contains arch, az, domain, hostname, localIP, os, realm, region. realm, region, az, domain were retrieved from environment variable named as REALM, REGION, AZ and DOMAIN. "*" means empty environment variable.|
| payloads | Contains RPC related metadata |
| error | Contains errors if occur |
| counters | Set by calling event.SetCounter() by user. |
| pairs | Set by calling event.AddPair() by user. |
| timing | Set by calling event.StartTimer() and event.EndTimer() by user. |
| remoteAddr |  As name described |
| operation | RPC method name |
| resCode | Response code of RPC |
| eventStatus | Ended or InProgress |

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

#### Metrics
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.interceptors.metricsProm.enabled | Enable metrics interceptor | boolean | false |

#### Auth
Enable the server side auth. codes.Unauthenticated would be returned to client if not authorized with user defined credential.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.interceptors.auth.enabled | Enable auth interceptor | boolean | false |
| gin.interceptors.auth.basic | Basic auth credentials as scheme of <user:pass> | []string | [] |
| gin.interceptors.auth.apiKey | API key auth | []string | [] |
| gin.interceptors.auth.ignorePrefix | The paths of prefix that will be ignored by interceptor | []string | [] |

#### Meta
Send application metadata as header to client.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.interceptors.meta.enabled | Enable meta interceptor | boolean | false |
| gin.interceptors.meta.prefix | Header key was formed as X-<Prefix>-XXX | string | RK |

#### Tracing
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| gin.interceptors.tracingTelemetry.enabled | Enable tracing interceptor | boolean | false |
| gin.interceptors.tracingTelemetry.exporter.file.enabled | Enable file exporter | boolean | RK |
| gin.interceptors.tracingTelemetry.exporter.file.outputPath | Export tracing info to files | string | stdout |
| gin.interceptors.tracingTelemetry.exporter.jaeger.agent.enabled | Export tracing info to jaeger agent | boolean | false |
| gin.interceptors.tracingTelemetry.exporter.jaeger.agent.host | As name described | string | localhost |
| gin.interceptors.tracingTelemetry.exporter.jaeger.agent.port | As name described | int | 6831 |
| gin.interceptors.tracingTelemetry.exporter.jaeger.collector.enabled | Export tracing info to jaeger collector | boolean | false |
| gin.interceptors.tracingTelemetry.exporter.jaeger.collector.endpoint | As name described | string | http://localhost:16368/api/trace |
| gin.interceptors.tracingTelemetry.exporter.jaeger.collector.username | As name described | string | "" |
| gin.interceptors.tracingTelemetry.exporter.jaeger.collector.password | As name described | string | "" |

### Development Status: Stable

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).


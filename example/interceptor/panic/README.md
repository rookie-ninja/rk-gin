# Panic middleware
In this example, we will try to create gin server with panic middleware enabled.

Panic interceptor will add do the bellow actions.
- Recover from panic
- Convert interface to standard rkerror.ErrorResp style of error
- Set resCode to 500
- Print stacktrace
- Set [panic:1] into event as counters
- Add error into event

**Please make sure panic interceptor to be added at last in chain of interceptors.**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Example](#example)
  - [Start server](#start-server)
  - [Output](#output)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-gin package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-gin
```
### Code
```go
import     "github.com/rookie-ninja/rk-gin/interceptor/panic"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
    interceptors := []gin.HandlerFunc{
        rkginpanic.Interceptor(),
    }
```

## Example
We will enable log interceptor to monitor RPC.

### Start server
```shell script
$ go run greeter-server.go
```

### Output
- Server side log (zap & event)
```shell script
2021-06-24T22:21:33.183+0800    ERROR   panic/interceptor.go:34 panic occurs:
goroutine 24 [running]:
...
main.Greeter(0xc0004fe100)
        /Users/dongxuny/workspace/rk/rk-gin/example/interceptor/panic/greeter-server.go:77 +0xa7
...
created by net/http.(*Server).Serve
        /usr/local/go/src/net/http/server.go:2969 +0x36c
        {"error": "[Internal Server Error] Panic manually!"}
```
```shell script
------------------------------------------------------------------------
endTime=2021-06-24T22:21:33.184343+08:00
startTime=2021-06-24T22:21:33.183715+08:00
elapsedNano=628133
timezone=CST
ids={"eventId":"adc9d7d5-cfc9-406e-9bef-3f8f3ea08d69"}
app={"appName":"rk","appVersion":"v0.0.0","entryName":"gin","entryType":"gin"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"10.8.0.2","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={"[Internal Server Error] Panic manually!":1}
counters={"panic":1}
pairs={}
timing={}
remoteAddr=localhost:62847
operation=/rk/v1/greeter
resCode=500
eventStatus=Ended
EOE
```
- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}
```

### Code
- [greeter-server.go](greeter-server.go)

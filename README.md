# rk-logger
Interceptor designed for gin framework.
Currently, supports bellow interceptors

- logging & metrics
- auth
- panic

## Installation
`go get -u github.com/rookie-ninja/rk-gin-interceptor`

## Quick Start
Interceptors can be used with chain.

### Logging & Metrics interceptor
Logging interceptor uses [zap logger](https://github.com/uber-go/zap) and [rk-query](https://github.com/rookie-ninja/rk-query) logs every requests.
[rk-prom](https://github.com/rookie-ninja/rk-prom) also used for prometheus metrics.

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	
	router.Use(
		rk_gin_inter_logging.RkGinZap(
			rk_gin_inter_logging.WithEventFactory(rk_query.NewEventFactory()),
			rk_gin_inter_logging.WithLogger(rk_logger.StdoutLogger)))

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})

	router.Run(":8080")
}
```

Output: 
```log
------------------------------------------------------------------------
end_time=2020-11-02T02:49:26.922327+08:00
start_time=2020-11-02T02:49:26.921547+08:00
time=0
hostname=JEREMYYIN-MB0
event_id=0307af32-4b76-4fb5-98a5-723659c87aa9
timing={}
counter={}
pair={}
error={}
field={"api.method":"GET","api.path":"/hello","api.protocol":"HTTP/1.1","api.query":"","app_version":"latest","az":"unknown","domain":"unknown","elapsed_ms":0,"end_time":"2020-11-02T02:49:26.922327+08:00","incoming_request_ids":[],"local.IP":"192.168.3.26","outgoing_request_id":["0307af32-4b76-4fb5-98a5-723659c87aa9"],"realm":"unknown","region":"unknown","remote.IP":"localhost","remote.port":"55403","res_code":200,"start_time":"2020-11-02T02:49:26.921547+08:00","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"}
remote_addr=localhost:55403
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
import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	"github.com/rookie-ninja/rk-gin-interceptor/panic/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		rk_gin_inter_logging.RkGinZap(
			rk_gin_inter_logging.WithEventFactory(rk_query.NewEventFactory()),
			rk_gin_inter_logging.WithLogger(rk_logger.StdoutLogger)),
		rk_gin_inter_panic.RkGinPanicZap())

	router.GET("/hello", func(ctx *gin.Context) {
		panic(errors.New("panic"))
		ctx.String(http.StatusOK, "Hello world")
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
import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/auth"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	"github.com/rookie-ninja/rk-gin-interceptor/panic/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		rk_gin_inter_logging.RkGinZap(
			rk_gin_inter_logging.WithEventFactory(rk_query.NewEventFactory()),
			rk_gin_inter_logging.WithLogger(rk_logger.StdoutLogger)),
		rk_gin_inter_auth.RkGinAuthZap(gin.Accounts{"user":"pass"}, "realm"),
		rk_gin_inter_panic.RkGinPanicZap())

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}
```
Output
```log
------------------------------------------------------------------------
end_time=2020-11-02T04:23:41.354791+08:00
start_time=2020-11-02T04:23:41.354758+08:00
time=0
hostname=JEREMYYIN-MB0
timing={}
counter={}
pair={}
error={}
field={"api.method":"GET","api.path":"/hello","api.protocol":"HTTP/1.1","api.query":"","app_version":"latest","az":"unknown","domain":"unknown","elapsed_ms":0,"end_time":"2020-11-02T04:23:41.354791+08:00","incoming_request_ids":[],"local.IP":"192.168.3.26","outgoing_request_id":[],"realm":"unknown","region":"unknown","remote.IP":"localhost","remote.port":"56698","res_code":401,"start_time":"2020-11-02T04:23:41.354758+08:00","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36"}
remote_addr=localhost:56698
app_name=Unknown
operation=GET-/hello
event_status=Ended
res_code=401
timezone=CST
os=darwin
arch=amd64
EOE
```

### Development Status: Stable

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
dongxuny@gmail.com. That email list is a private, safe space; even the zap
maintainers don't have access, so don't hesitate to hold us to a high
standard.

<hr>

Released under the [MIT License](LICENSE).


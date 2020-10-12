package rk_gin_inter_panic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// RkGinPanicZap returns a gin.HandlerFunc (middleware)
func RkGinPanicZap() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ed, exist := ctx.Get(rk_gin_inter_logging.RKEventKey)
		if !exist {
			if rk_gin_inter_logging.RKEventFactory == nil {
				rk_gin_inter_logging.RKEventFactory = rk_query.NewEventFactory()
			}

			ed = rk_gin_inter_logging.RKEventFactory.CreateEventNoop()
		}
		event := ed.(rk_query.Event)

		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
							event.AddErr(se)
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				if brokenPipe {
					event.AddFields(
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())))

					event.SetEndTime(time.Now())
					event.SetResCode(strconv.Itoa(http.StatusInternalServerError))
					// If the connection is dead, we can't write a status to it.
					ctx.Error(err.(error)) // nolint: errcheck
					ctx.Abort()
					return
				}

				event.AddFields(
					zap.String("request", string(httpRequest)),
					zap.String("stack", string(debug.Stack())))

				event.SetEndTime(time.Now())
				event.SetResCode(strconv.Itoa(http.StatusInternalServerError))

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}

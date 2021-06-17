// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginpanic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// PanicInterceptor returns a gin.HandlerFunc (middleware)
func PanicInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx == nil {
			return
		}

		event := rkginlog.GetEvent(ctx)

		defer func() {
			if err := recover(); err != nil {
				rkginlog.GetLogger(ctx).Error("panic occurs\n" + string(debug.Stack()))
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(
							strings.ToLower(se.Error()),
							"broken pipe") || strings.Contains(strings.ToLower(se.Error()),
							"connection reset by peer") {
							brokenPipe = true
							event.AddErr(se)
						}
					}
				}

				if brokenPipe {
					rkginlog.GetLogger(ctx).Error(string(debug.Stack()))
					event.SetEndTime(time.Now())
					event.SetResCode(strconv.Itoa(http.StatusInternalServerError))
					// If the connection is dead, we can't write a status to it.
					ctx.Error(err.(error)) // nolint: err check
					ctx.Abort()
					return
				}

				event.SetEndTime(time.Now())
				event.SetResCode(strconv.Itoa(http.StatusInternalServerError))

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}

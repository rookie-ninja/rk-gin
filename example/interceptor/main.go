// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rk_gin_log.RkGinLog(
			rk_gin_log.WithEventFactory(rk_query.NewEventFactory()),
			rk_gin_log.WithLogger(rk_logger.StdoutLogger)),
		rk_gin_auth.RkGinAuth(gin.Accounts{"user": "pass"}, "realm"),
		rk_gin_panic.RkGinPanic())

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}

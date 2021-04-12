// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/auth"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
)

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

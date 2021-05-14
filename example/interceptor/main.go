// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/gin-gonic/gin"
	rkginauth "github.com/rookie-ninja/rk-gin/interceptor/auth"
	rkginbasic "github.com/rookie-ninja/rk-gin/interceptor/basic"
	rkginlog "github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	rkginmetrics "github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	rkginpanic "github.com/rookie-ninja/rk-gin/interceptor/panic/zap"
	rklogger "github.com/rookie-ninja/rk-logger"
	rkquery "github.com/rookie-ninja/rk-query"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(
		rkginbasic.BasicInterceptor(),
		rkginlog.LoggingZapInterceptor(
			rkginlog.WithEventFactory(rkquery.NewEventFactory()),
			rkginlog.WithLogger(rklogger.StdoutLogger)),
		rkginmetrics.MetricsPromInterceptor(),
		rkginauth.BasicAuthInterceptor(gin.Accounts{"user": "pass"}, "realm"),
		rkginpanic.PanicInterceptor(),
	)

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
		//panic(errors.New(""))
	})
	router.Run(":8080")
}

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/context"
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
		rk_gin_inter_context.AddRequestIdToOutgoingHeader(ctx)
		rk_gin_inter_context.GetLogger(ctx).Info("This is info message")
		ctx.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}

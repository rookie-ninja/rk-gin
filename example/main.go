package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin-interceptor/logging/zap"
	"github.com/rookie-ninja/rk-gin-interceptor/panic/zap"
	"net/http"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(rk_gin_inter_logging.RkGinZap(nil), rk_gin_inter_panic.RkGinPanicZap())

	router.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world")
	})
	router.Run(":8080")
}

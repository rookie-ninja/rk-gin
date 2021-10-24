// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	rkginlimit "github.com/rookie-ninja/rk-gin/interceptor/ratelimit"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with rate limit interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ******************************************************
	// ********** Override App name and version *************
	// ******************************************************
	//
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "demo-app"
	// rkentry.GlobalAppCtx.GetAppInfoEntry().Version = "demo-version"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginlog.Interceptor(),
		rkginlimit.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkginmeta.WithEntryNameAndType("greeter", "gin"),
		//
		// Provide algorithm, rkgrpclimit.LeakyBucket and rkgrpclimit.TokenBucket was available, default is TokenBucket.
		//rkginlimit.WithAlgorithm(rkginlimit.LeakyBucket),
		//
		// Provide request per second, if provide value of zero, then no requests will be pass through and user will receive an error with
		// resource exhausted.
		//rkginlimit.WithReqPerSec(10),
		//
		// Provide request per second with path name.
		// The name should be full path name. if provide value of zero,
		// then no requests will be pass through and user will receive an error with resource exhausted.
		//rkginlimit.WithReqPerSecByPath("/rk/v1/greeter", 0),
		//
		// Provide user function of limiter
		//rkginlimit.WithGlobalLimiter(func(ctx *gin.Context) error {
		//	 return nil
		//}),
		//
		// Provide user function of limiter by path name.
		// The name should be full path name.
		//rkginlimit.WithLimiterByPath("/rk/v1/greeter", func(ctx *gin.Context) error {
		//	 return nil
		//}),
		),
	}

	// 1: Create gin server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start gin server.
func startGreeterServer(interceptors ...gin.HandlerFunc) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(interceptors...)
	router.GET("/rk/v1/greeter", Greeter)

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to listen: %v", err)
		}
	}()

	return server
}

// GreeterResponse Response of Greeter.
type GreeterResponse struct {
	Message string
}

// Greeter Handler.
func Greeter(ctx *gin.Context) {
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkginmeta.Interceptor() in server side.
	// 2: rkginctx.AddHeaderToClient(ctx, rkginctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkginctx.GetLogger(ctx).Info("Received request from client.")

	// Set request id with X-Request-Id to outgoing headers.
	// rkginctx.SetHeaderToClient(ctx, rkginctx.RequestIdKey, "this-is-my-request-id-overridden")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

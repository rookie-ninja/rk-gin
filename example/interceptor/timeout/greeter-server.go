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
	"github.com/rookie-ninja/rk-gin/interceptor/panic"
	"github.com/rookie-ninja/rk-gin/interceptor/timeout"
	"log"
	"net/http"
	"time"
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
		rkginpanic.Interceptor(),
		rkginlog.Interceptor(),
		rkgintimeout.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkgintimeout.WithEntryNameAndType("greeter", "gin"),
		//
		// Provide timeout and response handler, a default one would be assigned with http.StatusRequestTimeout
		// This option impact all routes
		// rkgintimeout.WithTimeoutAndResp(time.Second, nil),
		//
		// Provide timeout and response handler by path, a default one would be assigned with http.StatusRequestTimeout
		// rkgintimeout.WithTimeoutAndRespByPath("/rk/v1/healthy", time.Second, nil),
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

	// Sleep for 5 seconds waiting to be timed out by interceptor
	time.Sleep(10 * time.Second)

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

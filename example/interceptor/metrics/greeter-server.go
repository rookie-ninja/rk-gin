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
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-prom"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with metrics interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// Override app name which would replace namespace value in prometheus.
	rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "newApp"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginmetrics.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkginmetrics.WithEntryNameAndType("greeter", "grpc"),
		//
		// Provide new prometheus registerer.
		// Default value is prometheus.DefaultRegisterer
		// rkginmetrics.WithRegisterer(prometheus.NewRegistry()),
		),
	}

	// 1: Start prometheus client
	// By default, we will start prometheus client at localhost:1608/metrics
	promEntry := rkprom.RegisterPromEntry()
	promEntry.Bootstrap(context.Background())
	defer promEntry.Interrupt(context.Background())

	// 2: Create gin server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 3: Wait for ctrl-C to shutdown server
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

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

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
	rkginctx "github.com/rookie-ninja/rk-gin/interceptor/context"
	rkginlog "github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	rkgintrace "github.com/rookie-ninja/rk-gin/interceptor/tracing/telemetry"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with tracing interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ****************************************
	// ********** Create Exporter *************
	// ****************************************

	// Export trace to stdout
	exporter := rkgintrace.CreateFileExporter("stdout")

	// Export trace to local file system
	// exporter := rkgintrace.CreateFileExporter("logs/trace.log")

	// Export trace to jaeger collector
	// exporter := rkgintrace.CreateJaegerExporter("localhost:14268", "", "")

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginlog.Interceptor(),
		rkgintrace.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkgintrace.WithEntryNameAndType("greeter", "grpc"),
			//
			// Provide an exporter.
			rkgintrace.WithExporter(exporter),
		//
		// Provide propagation.TextMapPropagator
		// rkgintrace.WithPropagator(<propagator>),
		//
		// Provide SpanProcessor
		// rkgintrace.WithSpanProcessor(<span processor>),
		//
		// Provide TracerProvider
		// rkgintrace.WithTracerProvider(<trace provider>),
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

// Response.
type GreeterResponse struct {
	Message string
}

// Handler.
func Greeter(ctx *gin.Context) {
	rkginctx.GetLogger(ctx).Info("Received request from client.")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

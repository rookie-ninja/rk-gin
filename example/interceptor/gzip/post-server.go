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
	"github.com/rookie-ninja/rk-gin/interceptor/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

// In this example, we will start a new gin server with gzip interceptor enabled.
// Listen on port of 8080 with POST /rk/v1/post.
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
		rkgingzip.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkgingzip.WithEntryNameAndType("greeter", "echo"),
		//
		// Provide level of compression.
		// Available options are
		// - NoCompression
		// - BestSpeed
		// - BestCompression
		// - DefaultCompression
		// - HuffmanOnly
		//rkgingzip.WithLevel(rkgingzip.DefaultCompression),
		//
		// Provide skipper function
		//rkgingzip.WithSkipper(func(e echo.Context) bool {
		//	return false
		//}),
		),
	}

	// 1: Create gin server
	server := startPostServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start gin server.
func startPostServer(interceptors ...gin.HandlerFunc) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(interceptors...)
	router.POST("/rk/v1/post", post)

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

// PostResponse Response of Greeter.
type PostResponse struct {
	ReceivedMessage string
}

// post Handler.
func post(ctx *gin.Context) {
	buf := new(strings.Builder)
	io.Copy(buf, ctx.Request.Body)

	ctx.JSON(http.StatusOK, &PostResponse{
		ReceivedMessage: fmt.Sprintf("Received %s!", buf.String()),
	})
}

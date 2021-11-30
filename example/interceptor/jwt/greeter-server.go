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
	"github.com/rookie-ninja/rk-gin/interceptor/jwt"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with jwt interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter.
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
		rkginjwt.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkginjwt.WithEntryNameAndType("greeter", "gin"),
			//
			// Required, provide signing key.
			rkginjwt.WithSigningKey([]byte("my-secret")),
			//
			// Optional, provide skipper function
			//rkginjwt.WithSkipper(func(e *gin.Context) bool {
			//	return true
			//}),
			//
			// Optional, provide token parse function, default one will be assigned.
			//rkginjwt.WithParseTokenFunc(func(auth string, ctx *gin.Context) (*jwt.Token, error) {
			//	return nil, nil
			//}),
			//
			// Optional, provide key function, default one will be assigned.
			//rkginjwt.WithKeyFunc(func(token *jwt.Token) (interface{}, error) {
			//	return nil, nil
			//}),
			//
			// Optional, default is Bearer
			//rkginjwt.WithAuthScheme("Bearer"),
			//
			// Optional
			//rkginjwt.WithTokenLookup("header:my-jwt-header-key"),
			//
			// Optional, default is HS256
			//rkginjwt.WithSigningAlgorithm(rkginjwt.AlgorithmHS256),
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
	rkginctx.GetLogger(ctx).Info("Received request from client.")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Is token valid:%v!", rkginctx.GetJwtToken(ctx)),
	})
}

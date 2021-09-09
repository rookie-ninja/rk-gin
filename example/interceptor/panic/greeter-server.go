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
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/panic"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with panic interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginlog.Interceptor(),
		rkginpanic.Interceptor(),
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
	// All bellow panic case should return same error response.
	// {"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}

	// Panic interceptor will wrap error with standard RK style error.
	// Please refer to rkerror.ErrorResp.
	// panic(errors.New("Panic manually!"))

	// Please refer to rkerror.ErrorResp.
	// panic(rkerror.FromError(errors.New("Panic manually!")))

	// Please refer to rkerror.ErrorResp.
	panic("Panic manually!")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

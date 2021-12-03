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
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"net/http"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/csrf/boot.yaml")

	// Bootstrap echo entry from boot config
	res := rkgin.RegisterGinEntriesWithConfig("example/boot/csrf/boot.yaml")

	// Register GET and POST method of /rk/v1/greeter
	entry := res["greeter"].(*rkgin.GinEntry)
	entry.Router.GET("/rk/v1/greeter", Greeter)
	entry.Router.POST("/rk/v1/greeter", Greeter)

	// Bootstrap echo entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	res["greeter"].Interrupt(context.Background())
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
		Message: fmt.Sprintf("CSRF token:%v", rkginctx.GetCsrfToken(ctx)),
	})
}

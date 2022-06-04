// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-gin/v2/boot"
	"github.com/rookie-ninja/rk-gin/v2/middleware/context"
	"net/http"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapBuiltInEntryFromYAML(boot)
	rkentry.BootstrapPluginEntryFromYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntryYAML(boot)

	// Register GET and POST method of /rk/v1/greeter
	ginEntry := res["greeter"].(*rkgin.GinEntry)
	ginEntry.Router.GET("/rk/v1/greeter", Greeter)
	ginEntry.Router.POST("/rk/v1/greeter", Greeter)

	// Bootstrap gin entry
	ginEntry.Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	ginEntry.Interrupt(context.Background())
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

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
	"net/http"
)

//go:embed boot.yaml
var boot []byte

////go:embed docs
//var docsFS embed.FS
//
////go:embed docs
//var staticFS embed.FS

func init() {
	//rkentry.GlobalAppCtx.AddEmbedFS(rkentry.DocsEntryType, "greeter", &docsFS)
	//rkentry.GlobalAppCtx.AddEmbedFS(rkentry.SWEntryType, "greeter", &docsFS)
	//rkentry.GlobalAppCtx.AddEmbedFS(rkentry.StaticFileHandlerEntryType, "greeter", &staticFS)
}

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntryYAML(boot)

	// Get GinEntry
	ginEntry := res["greeter"].(*rkgin.GinEntry)
	ginEntry.Router.GET("/v1/greeter", Greeter)

	// Bootstrap gin entry
	ginEntry.Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	ginEntry.Interrupt(context.Background())
}

func Greeter(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

type GreeterResponse struct {
	Message string
}

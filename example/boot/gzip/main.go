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
	"io"
	"net/http"
	"strings"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapBuiltInEntryFromYAML(boot)
	rkentry.BootstrapPluginEntryFromYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntryYAML(boot)
	ginEntry := res["greeter"]

	// Bootstrap gin entry
	ginEntry.Bootstrap(context.Background())

	// Register post method
	ginEntry.(*rkgin.GinEntry).Router.POST("/rk/v1/post", post)

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	ginEntry.Interrupt(context.Background())
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

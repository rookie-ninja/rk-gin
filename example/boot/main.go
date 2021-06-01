// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-gin/interceptor/extension"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
)

func main() {
	bootFromConfig()
	//bootFromCode()
}

func bootFromConfig() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/boot.yaml")

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntriesWithConfig("example/boot/boot.yaml")

	// Bootstrap gin entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	res["greeter"].Interrupt(context.Background())
}

func bootFromCode() {
	// Create gin entry
	entry := rkgin.RegisterGinEntry(
		rkgin.WithNameGin("greeter"),
		rkgin.WithPortGin(8080),
		rkgin.WithCommonServiceEntryGin(rkgin.NewCommonServiceEntry()),
		rkgin.WithInterceptorsGin(rkginlog.LoggingZapInterceptor([]rkginlog.Option{}...),
			rkginextension.ExtensionInterceptor()))

	// Start server
	go entry.Bootstrap(context.Background())

	// Wait for shutdown sig
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt server
	entry.Interrupt(context.Background())
}

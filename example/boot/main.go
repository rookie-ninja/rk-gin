// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
)

func main() {
	bootFromConfig()
}

func bootFromConfig() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterBasicEntriesFromConfig("example/boot/boot.yaml")

	// Bootstrap gin entry from boot config
	res := rkgin.RegisterGinEntriesWithConfig("example/boot/boot.yaml")

	// Bootstrap gin entry
	go res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	res["greeter"].Interrupt(context.Background())
}

func bootFromCode() {
	// Create event data
	fac := rkquery.NewEventFactory()

	// Create options for interceptor
	opts := []rkginlog.Option{
		rkginlog.WithEventFactory(fac),
		rkginlog.WithLogger(rklogger.StdoutLogger),
	}

	// Create gin entry
	entry := rkgin.RegisterGinEntry(
		rkgin.WithNameGin("greeter"),
		rkgin.WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		rkgin.WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()),
		rkgin.WithPortGin(8080),
		rkgin.WithCommonServiceEntryGin(rkgin.NewCommonServiceEntry()),
		rkgin.WithTVEntryGin(rkgin.NewTVEntry()),
		rkgin.WithInterceptorsGin(rkginlog.LoggingZapInterceptor(opts...)))

	// Start server
	go entry.Bootstrap(context.Background())

	// Wait for shutdown sig
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt server
	entry.Interrupt(context.Background())
}

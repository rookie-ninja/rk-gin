// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"time"
)

func main() {
	// create event data
	fac := rk_query.NewEventFactory()

	// create options for interceptor
	opts := []rk_gin_log.Option{
		rk_gin_log.WithEventFactory(fac),
		rk_gin_log.WithLogger(rk_logger.StdoutLogger),
		rk_gin_log.WithEnableLogging(true),
		rk_gin_log.WithEnableMetrics(true),
	}

	// create gin entry
	entry := rk_gin.NewGinEntry(
		rk_gin.WithName("greeter"),
		rk_gin.WithEventFactory(fac),
		rk_gin.WithLogger(rk_logger.StdoutLogger),
		rk_gin.WithPort(8080),
		rk_gin.WithEnableCommonService(true),
		rk_gin.WithInterceptors(rk_gin_log.RkGinLog(opts...)))

	// start server
	entry.Bootstrap(fac.CreateEvent())
	entry.Wait(1 * time.Second)
}

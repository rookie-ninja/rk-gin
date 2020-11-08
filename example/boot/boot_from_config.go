// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
)

func main() {
	fac := rk_query.NewEventFactory()
	entries := rk_gin.NewGinEntries("example/boot/boot.yaml", fac, rk_logger.StdoutLogger)
	entries["greeter"].Bootstrap(fac.CreateEvent())
}

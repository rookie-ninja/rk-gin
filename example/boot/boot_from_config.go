// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-gin/boot"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"time"
)

// @title RK Swagger Example
// @version 1.0
// @description This is a sample demo server with rk-gin.
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.basic BasicAuth

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	fac := rk_query.NewEventFactory()
	entries := rk_gin.NewGinEntries("example/boot/boot.yaml", fac, rk_logger.StdoutLogger)
	entries["greeter"].Bootstrap(fac.CreateEvent())
	entries["greeter"].Wait(1 * time.Second)
}

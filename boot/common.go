// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	rk_info "github.com/rookie-ninja/rk-common/info"
	rk_gin_ctx "github.com/rookie-ninja/rk-gin/interceptor/context"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
)

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func shutdownWithError(err error) {
	debug.PrintStack()
	glog.Error(err)
	os.Exit(1)
}

// common services
func healthy(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	event.AddPair("healthy", "true")

	ctx.JSON(http.StatusOK, gin.H{
		"healthy": true,
	})
}

func gc(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	before := rk_info.MemStatsToStruct()
	runtime.GC()
	event.AddFields(rk_info.MemStatsToFields()...)
	after := rk_info.MemStatsToStruct()

	ctx.JSON(http.StatusOK, gin.H{
		"mem_stat_before_gc": before,
		"mem_stat_after_gc":  after,
	})
}

func info(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)
	event := rk_gin_ctx.GetEvent(ctx)

	event.AddFields(rk_info.BasicInfoToFields()...)

	ctx.JSON(http.StatusOK, gin.H{
		"info": rk_info.BasicInfoToStruct(),
	})
}

func dumpConfig(ctx *gin.Context) {
	// Add auto generated request ID
	rk_gin_ctx.AddRequestIdToOutgoingHeader(ctx)

	ctx.JSON(http.StatusOK, gin.H{
		"viper": rk_info.ViperConfigToJSON(),
		"rk":    rk_info.RkConfigToJSON(),
	})
}

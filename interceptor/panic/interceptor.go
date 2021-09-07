// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkginpanic is a middleware of gin framework for recovering from panic
package rkginpanic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	rkgininter "github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// Interceptor returns a gin.HandlerFunc (middleware)
func Interceptor(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkgininter.RpcEntryNameKey, set.EntryName)

		defer func() {
			if recv := recover(); recv != nil {
				var res *rkerror.ErrorResp

				if se, ok := recv.(*rkerror.ErrorResp); ok {
					res = se
				} else if re, ok := recv.(error); ok {
					res = rkerror.FromError(re)
				} else {
					res = rkerror.New(rkerror.WithMessage(fmt.Sprintf("%v", recv)))
				}

				rkginctx.GetEvent(ctx).SetCounter("panic", 1)
				rkginctx.GetEvent(ctx).AddErr(res.Err)
				rkginctx.GetLogger(ctx).Error(fmt.Sprintf("panic occurs:\n%s", string(debug.Stack())), zap.Error(res.Err))

				ctx.JSON(http.StatusInternalServerError, res)
			}
		}()

		ctx.Next()
	}
}

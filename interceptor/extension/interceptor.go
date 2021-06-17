// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginextension

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/basic"
	"time"
)

const (
	RequestIdHeaderKeyDefault = "X-Request-Id"
	TraceIdHeaderKeyDefault   = "X-Trace-Id"
)

var (

)

// Add common headers as extension style in http response.
// The key is defined as bellow:
// 1: X-Request-Id: Request id generated by interceptor.
// 2: X-Trace-Id: Trace id generated by interceptor.
// 3: X-<Prefix>-Location: A valid URI.
// 4: X-<Prefix-App: Application name.
// 5: X-<Prefix>-App-Version: Version of application.
// 6: X-<Prefix>-App-Unix-Time: Unix time of current application.
// 7: X-<Prefix>-Request-Received-Time: Time of current request received by application.
func ExtensionInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName:    rkginbasic.RkEntryNameValue,
		EntryType:    rkginbasic.RkEntryTypeValue,
		Prefix:       "RK",
		RequestIdKey: "X-RK-Trace-Id",
		TraceIdKey:   "X-RK-Trace-Id",
	}

	for i := range opts {
		opts[i](set)
	}

	if len(set.Prefix) < 1 {
		set.Prefix = "RK"
	}

	set.TraceIdKey = TraceIdHeaderKeyDefault
	set.RequestIdKey = RequestIdHeaderKeyDefault
	set.AppNameKey = fmt.Sprintf("X-%s-App-Name", set.Prefix)
	set.AppNameValue = rkentry.GlobalAppCtx.GetAppInfoEntry().AppName
	set.AppVersionKey = fmt.Sprintf("X-%s-App-Version", set.Prefix)
	set.AppVersionValue = rkentry.GlobalAppCtx.GetAppInfoEntry().Version
	set.AppUnixTimeKey = fmt.Sprintf("X-%s-App-Unix-Time", set.Prefix)
	set.ReceivedTimeKey = fmt.Sprintf("X-%s-Received-Time", set.Prefix)

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return func(ctx *gin.Context) {
		requestId := rkcommon.GenerateRequestId()
		ctx.Header(set.RequestIdKey, requestId)

		ctx.Header(set.AppNameKey, set.AppNameValue)
		ctx.Header(set.AppVersionKey, set.AppVersionValue)

		now := time.Now()
		ctx.Header(set.AppUnixTimeKey, now.Format(time.RFC3339Nano))
		ctx.Header(set.ReceivedTimeKey, now.Format(time.RFC3339Nano))

		ctx.Next()
	}
}

// Get option set with gin.Context
func GetOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkginbasic.RkEntryNameKey)
	return optionsMap[entryName]
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing extension interceptor
type optionSet struct {
	EntryName       string
	EntryType       string
	Prefix          string
	RequestIdKey    string
	TraceIdKey      string
	AppNameKey      string
	AppNameValue    string
	AppVersionKey   string
	AppVersionValue string
	AppUnixTimeKey  string
	ReceivedTimeKey string
}

type Option func(*optionSet)

// Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// Provide prefix of response header as bellow:
// X-<Prefix>-XXX
func WithPrefix(prefix string) Option {
	return func(opt *optionSet) {
		opt.Prefix = prefix
	}
}

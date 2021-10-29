// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintimeout

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-gin/interceptor"
	rkginctx "github.com/rookie-ninja/rk-gin/interceptor/context"
	"net/http"
	"strings"
	"time"
)

const global = "rk-global"

var (
	defaultResponse = func(ctx *gin.Context) {
		ctx.JSON(http.StatusRequestTimeout, rkerror.New(
			rkerror.WithHttpCode(http.StatusRequestTimeout),
			rkerror.WithMessage("Request timed out!")))
	}
	defaultTimeout  = 5 * time.Second
	globalTimeoutRk = &timeoutRk{
		timeout:  defaultTimeout,
		response: defaultResponse,
	}
)

type timeoutRk struct {
	timeout  time.Duration
	response gin.HandlerFunc
}

// Interceptor would distinguish auth set based on.
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName: rkgininter.RpcEntryNameValue,
		EntryType: rkgininter.RpcEntryTypeValue,
		timeouts:  make(map[string]*timeoutRk),
	}

	for i := range opts {
		opts[i](set)
	}

	// add global timeout
	set.timeouts[global] = &timeoutRk{
		timeout:  globalTimeoutRk.timeout,
		response: globalTimeoutRk.response,
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Tick will continue request it rest of handlers.
// If timeout triggered, then return http.StatusRequestTimeout back to client.
//
// Mainly copied from https://github.com/gin-contrib/timeout/blob/master/timeout.go
func (set *optionSet) Tick(ctx *gin.Context, path string) {
	rk := set.getTimeoutRk(path)

	event := rkginctx.GetEvent(ctx)

	// 1: create three channels
	//
	// finishChan: triggered while request has been handled successfully
	// panicChan: triggered while panic occurs
	// timeoutChan: triggered while timing out
	finishChan := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)
	timeoutChan := time.After(rk.timeout)

	// 2: create a buffer pool and new writer
	// Why?
	//
	// We may face the case that request timed out while user code is writing to response writer.
	// So, we create a new writer with mutex lock and ignore contents user code writers if timed out .
	bufPool := &bufferPool{}
	buffer := bufPool.Get()
	originalWriter := ctx.Writer
	newWriter := newWriter(originalWriter, buffer)

	// 3: assign new writer
	ctx.Writer = newWriter

	// 4: create a new go routine catch panic
	go func() {
		defer func() {
			if recv := recover(); recv != nil {
				panicChan <- recv
				fmt.Println("I am herer")
			}
		}()

		ctx.Next()
		finishChan <- struct{}{}
	}()

	// 5: waiting for three channels
	select {
	// 5.1: switch to original writer and panic
	case recv := <-panicChan:
		newWriter.FreeBuffer()
		ctx.Writer = originalWriter
		panic(recv)
	// 5.2: copy headers and contents into original writer
	case <-finishChan:
		newWriter.mu.Lock()
		defer newWriter.mu.Unlock()

		// copy headers and code
		dst := newWriter.ResponseWriter.Header()
		for k, vv := range newWriter.Header() {
			dst[k] = vv
		}
		newWriter.ResponseWriter.WriteHeader(newWriter.code)

		// copy contents
		if _, err := newWriter.ResponseWriter.Write(buffer.Bytes()); err != nil {
			panic(err)
		}

		// free buffer
		newWriter.FreeBuffer()
		bufPool.Put(buffer)
	// 5.3: free buffer and switch to original writer. Timeout response will be attached to response too.
	case <-timeoutChan:
		ctx.Abort()

		// set as timeout
		event.SetCounter("timeout", 1)

		newWriter.mu.Lock()
		defer newWriter.mu.Unlock()

		newWriter.timeout = true

		// free buffer
		newWriter.FreeBuffer()
		bufPool.Put(buffer)

		// switch to original writer
		ctx.Writer = originalWriter

		// write timed out response
		rk.response(ctx)

		// switch back to new writer since user code may still want to write to it.
		// Panic may occur if we ignore this step.
		ctx.Writer = newWriter
	}
}

// Get timeout instance with path.
// Global one will be returned if no not found.
func (set *optionSet) getTimeoutRk(path string) *timeoutRk {
	if v, ok := set.timeouts[path]; ok {
		return v
	}

	return set.timeouts[global]
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	EntryName string
	EntryType string
	timeouts  map[string]*timeoutRk
}

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// WithTimeoutAndResp Provide global timeout and response handler.
// If response is nil, default globalResponse will be assigned
func WithTimeoutAndResp(timeout time.Duration, resp gin.HandlerFunc) Option {
	return func(set *optionSet) {
		if resp == nil {
			resp = defaultResponse
		}

		if timeout == 0 {
			timeout = defaultTimeout
		}

		globalTimeoutRk.timeout = timeout
		globalTimeoutRk.response = resp
	}
}

// WithTimeoutAndRespByPath Provide timeout and response handler by path.
// If response is nil, default globalResponse will be assigned
func WithTimeoutAndRespByPath(path string, timeout time.Duration, resp gin.HandlerFunc) Option {
	return func(set *optionSet) {
		path = normalisePath(path)

		if resp == nil {
			resp = defaultResponse
		}

		if timeout == 0 {
			timeout = defaultTimeout
		}

		set.timeouts[path] = &timeoutRk{
			timeout:  timeout,
			response: resp,
		}
	}
}

func normalisePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

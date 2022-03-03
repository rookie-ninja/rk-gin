// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Copied from https://github.com/gin-contrib/timeout/blob/master/writer.go

package rkgintout

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// writer is a writer with memory buffer
type writer struct {
	gin.ResponseWriter
	body         *bytes.Buffer
	headers      http.Header
	mu           sync.Mutex
	timeout      bool
	wroteHeaders bool
	code         int
}

// newWriter will return a timeout.Writer pointer
func newWriter(w gin.ResponseWriter, buf *bytes.Buffer) *writer {
	return &writer{ResponseWriter: w, body: buf, headers: make(http.Header)}
}

// Write will write data to response body
func (w *writer) Write(data []byte) (int, error) {
	if w.timeout || w.body == nil {
		return 0, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	return w.body.Write(data)
}

// WriteHeader will write http status code
func (w *writer) WriteHeader(code int) {
	checkWriteHeaderCode(code)
	if w.timeout || w.wroteHeaders {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.writeHeader(code)
}

func (w *writer) writeHeader(code int) {
	w.wroteHeaders = true
	w.code = code
}

// Header will get response headers
func (w *writer) Header() http.Header {
	return w.headers
}

// WriteString will write string to response body
func (w *writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// FreeBuffer will release buffer pointer
func (w *writer) FreeBuffer() {
	w.body = nil
}

func checkWriteHeaderCode(code int) {
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid http status code: %d", code))
	}
}

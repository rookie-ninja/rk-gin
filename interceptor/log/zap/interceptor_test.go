// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginlog

import (
	"errors"
	"github.com/gin-gonic/gin"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	rkginctx "github.com/rookie-ninja/rk-gin/interceptor/context"
	rkquery "github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		data:   make([]byte, 0),
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	data       []byte
	statusCode int
	header     http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	m.data = bytes
	return len(bytes), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestInterceptor_WithShouldNotLog(t *testing.T) {
	defer assertNotPanic(t)
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "/rk/v1/assets",
		},
	}

	handler(ctx)
}

func TestInterceptor_HappyCase(t *testing.T) {
	defer assertNotPanic(t)
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
	}

	ctx.Error(errors.New("ut error"))
	ctx.Writer.Header().Set(rkginctx.RequestIdKey, "ut-request-id")
	ctx.Writer.Header().Set(rkginctx.TraceIdKey, "ut-trace-id")

	handler(ctx)

	event := rkginctx.GetEvent(ctx)
	assert.NotEmpty(t, event.GetRemoteAddr())
	assert.NotEmpty(t, event.ListPayloads())
	assert.NotEmpty(t, event.GetOperation())
	assert.NotZero(t, event.GetErrCount(errors.New("ut error")))
	assert.NotEmpty(t, event.GetRequestId())
	assert.NotEmpty(t, event.GetTraceId())
	assert.NotEmpty(t, event.GetResCode())
	assert.Equal(t, rkquery.Ended, event.GetEventStatus())

	assert.False(t, ctx.IsAborted())
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

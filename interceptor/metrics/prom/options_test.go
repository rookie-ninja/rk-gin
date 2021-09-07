// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	rkgininter "github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestWithEntryNameAndType(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)

	defer clearAllMetrics()
}

func TestWithRegisterer(t *testing.T) {
	reg := prometheus.NewRegistry()
	set := newOptionSet(
		WithRegisterer(reg))

	assert.Equal(t, reg, set.registerer)

	defer clearAllMetrics()
}

func TestGetOptionSet(t *testing.T) {
	// With nil context
	assert.Nil(t, getOptionSet(nil))

	// Happy case
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")
	set := newOptionSet()
	optionsMap["ut-entry"] = set
	assert.Equal(t, set, getOptionSet(ctx))

	defer clearAllMetrics()
}

func TestGetServerMetricsSet(t *testing.T) {
	reg := prometheus.NewRegistry()
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")
	assert.Equal(t, set.MetricsSet, GetServerMetricsSet(ctx))

	defer clearAllMetrics()
}

func TestListServerMetricsSets(t *testing.T) {
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")
	assert.Len(t, ListServerMetricsSets(), 1)

	defer clearAllMetrics()
}

func TestGetServerResCodeMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerResCodeMetrics(nil))

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	defer clearAllMetrics()
}

func TestGetServerErrorMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerErrorMetrics(nil))

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	defer clearAllMetrics()
}

func TestGetServerDurationMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerDurationMetrics(nil))

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Set(rkgininter.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerDurationMetrics(ctx))

	defer clearAllMetrics()
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintrace

import (
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"testing"
)

func TestWithEntryNameAndType(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
}

func TestWithExporter(t *testing.T) {
	exporter := &NoopExporter{}
	set := newOptionSet(
		WithExporter(exporter))

	assert.Equal(t, exporter, set.Exporter)
}

func TestWithSpanProcessor(t *testing.T) {
	processor := sdktrace.NewSimpleSpanProcessor(&NoopExporter{})
	set := newOptionSet(
		WithSpanProcessor(processor))

	assert.Equal(t, processor, set.Processor)
}

func TestWithTracerProvider(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	set := newOptionSet(
		WithTracerProvider(provider))

	assert.Equal(t, provider, set.Provider)
}

func TestWithPropagator(t *testing.T) {
	prop := propagation.NewCompositeTextMapPropagator()
	set := newOptionSet(
		WithPropagator(prop))

	assert.Equal(t, prop, set.Propagator)
}

func TestShutdownExporters(t *testing.T) {
	assertNotPanic(t)
	newOptionSet(
		WithExporter(&NoopExporter{}))

	ShutdownExporters()
}

func TestNoopExporter_ExportSpans(t *testing.T) {
	exporter := NoopExporter{}
	assert.Nil(t, exporter.ExportSpans(nil, nil))
}

func TestNoopExporter_Shutdown(t *testing.T) {
	exporter := NoopExporter{}
	assert.Nil(t, exporter.Shutdown(nil))
}

func TestCreateNoopExporter(t *testing.T) {
	assert.NotNil(t, CreateNoopExporter())
}

func TestCreateJaegerExporter(t *testing.T) {
	// without endpoint
	exporter := CreateJaegerExporter("", "ut-user", "ut-pass")
	assert.NotNil(t, exporter)

	// with endpoint
	exporter = CreateJaegerExporter("localhost:14268", "ut-user", "ut-pass")
	assert.NotNil(t, exporter)
}

func TestCreateFileExporter(t *testing.T) {
	// with stdout
	exporter := CreateFileExporter("")
	assert.NotNil(t, exporter)

	// with non stdout
	exporter = CreateFileExporter("stderror")
	assert.NotNil(t, exporter)
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

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgintrace

import (
	rkginbasic "github.com/rookie-ninja/rk-gin/interceptor/basic"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"testing"
)

func TestWithExporter_HappyCase(t *testing.T) {
	set := &optionSet{}

	exporter, _ := stdout.NewExporter()
	opt := WithExporter(exporter)

	opt(set)

	assert.Equal(t, exporter, set.Exporter)
}

func TestWithSpanProcessor_HappyCase(t *testing.T) {
	set := &optionSet{}

	processor := sdktrace.NewBatchSpanProcessor(nil)
	opt := WithSpanProcessor(processor)

	opt(set)

	assert.Equal(t, processor, set.Processor)
}

func TestWithTracerProvider_HappyCase(t *testing.T) {
	set := &optionSet{}

	provider := sdktrace.NewTracerProvider()
	opt := WithTracerProvider(provider)

	opt(set)

	assert.Equal(t, provider, set.Provider)
}

func TestWithPropagator_HappyCase(t *testing.T) {
	set := &optionSet{}

	propagator := propagation.NewCompositeTextMapPropagator()
	opt := WithPropagator(propagator)

	opt(set)

	assert.Equal(t, propagator, set.Propagator)
}

func TestWithEntryNameAndType_HappyCase(t *testing.T) {
	set := &optionSet{}

	entryName, entryType := "ut-entry-name", "ut-entry"
	opt := WithEntryNameAndType(entryName, entryType)

	opt(set)

	assert.Equal(t, entryName, set.EntryName)
	assert.Equal(t, entryType, set.EntryType)
}

func TestCreateFileExporter_WithEmptyOutputPath(t *testing.T) {
	assert.NotNil(t, CreateFileExporter(""))
}

func TestCreateFileExporter_WithStdout(t *testing.T) {
	assert.NotNil(t, CreateFileExporter("stdout"))
}

func TestCreateFileExporter_HappyCase(t *testing.T) {
	assert.NotNil(t, CreateFileExporter("logs/trace.log"))
}

func TestTelemetryInterceptor_WithoutOption(t *testing.T) {
	TelemetryInterceptor()

	set := optionsMap[rkginbasic.RkEntryNameValue]
	assert.NotNil(t, set)
	assert.NotNil(t, set.Propagator)
	assert.NotNil(t, set.Provider)
	assert.NotNil(t, set.Processor)
	assert.NotNil(t, set.Exporter)
	assert.NotNil(t, set.Tracer)
	assert.Equal(t, rkginbasic.RkEntryNameValue, set.EntryName)
	assert.Equal(t, rkginbasic.RkEntryTypeValue, set.EntryType)
}

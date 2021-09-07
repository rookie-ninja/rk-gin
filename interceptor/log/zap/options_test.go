// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginlog

import (
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithEntryNameAndType(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
}

func TestWithZapLoggerEntry(t *testing.T) {
	entry := rkentry.NoopZapLoggerEntry()
	set := newOptionSet(
		WithZapLoggerEntry(entry))
	assert.Equal(t, entry, set.zapLoggerEntry)
}

func TestWithEventLoggerEntry(t *testing.T) {
	entry := rkentry.NoopEventLoggerEntry()
	set := newOptionSet(
		WithEventLoggerEntry(entry))
	assert.Equal(t, entry, set.eventLoggerEntry)
}

func TestWithZapLoggerEncoding(t *testing.T) {
	set := newOptionSet(
		WithZapLoggerEncoding(ENCODING_JSON))

	assert.Equal(t, ENCODING_JSON, set.zapLoggerEncoding)
}

func TestWithZapLoggerOutputPaths(t *testing.T) {
	set := newOptionSet(
		WithZapLoggerOutputPaths("ut-path"))

	assert.Contains(t, set.zapLoggerOutputPath, "ut-path")
}

func TestWithEventLoggerEncoding(t *testing.T) {
	// Test with console encoding
	set := newOptionSet(
		WithEventLoggerEncoding(ENCODING_CONSOLE))
	assert.Equal(t, rkquery.CONSOLE, set.eventLoggerEncoding)

	// Test with json encoding
	set = newOptionSet(
		WithEventLoggerEncoding(ENCODING_JSON))
	assert.Equal(t, rkquery.JSON, set.eventLoggerEncoding)

	// Test with non console and json
	set = newOptionSet(
		WithEventLoggerEncoding(-1))
	assert.Equal(t, rkquery.CONSOLE, set.eventLoggerEncoding)
}

func TestWithEventLoggerOutputPaths(t *testing.T) {
	set := newOptionSet(
		WithEventLoggerOutputPaths("ut-path"))
	assert.Contains(t, set.eventLoggerOutputPath, "ut-path")
}

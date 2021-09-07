// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginauth

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithEntryNameAndType(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
}

func TestWithBasicAuth(t *testing.T) {
	set := newOptionSet(
		WithBasicAuth("ut-realm", "user:pass"))

	assert.Equal(t, "ut-realm", set.BasicRealm)
	assert.True(t, set.BasicAccounts[base64.StdEncoding.EncodeToString([]byte("user:pass"))])
}

func TestWithApiKeyAuth(t *testing.T) {
	set := newOptionSet(
		WithApiKeyAuth("ut-api-key"))

	assert.True(t, set.ApiKey["ut-api-key"])
}

func TestWithIgnorePrefix(t *testing.T) {
	set := newOptionSet(
		WithIgnorePrefix("ut-prefix"))

	assert.Contains(t, set.IgnorePrefix, "ut-prefix")
}

func TestOptionSet_Authorized(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	// With invalid auth type
	assert.False(t, set.Authorized("invalid", ""))

	// With invalid basic auth
	assert.False(t, set.Authorized(typeBasic, "invalid"))
	// With valid basic auth
	assert.True(t, set.Authorized(typeBasic, base64.StdEncoding.EncodeToString([]byte("user:pass"))))

	// With invalid api key
	assert.False(t, set.Authorized(typeApiKey, "invalid"))
	// With valid api key
	assert.True(t, set.Authorized(typeApiKey, "ut-api-key"))
}

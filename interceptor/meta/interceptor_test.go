// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkginmeta

import (
	"github.com/rookie-ninja/rk-gin/interceptor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithEntryNameAndType_HappyCase(t *testing.T) {
	opt := WithEntryNameAndType("ut-name", "ut-type")

	set := &optionSet{}

	opt(set)

	assert.Equal(t, "ut-name", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
}

func TestWithPrefix_HappyCase(t *testing.T) {
	opt := WithPrefix("ut-prefix")

	set := &optionSet{}

	opt(set)

	assert.Equal(t, "ut-prefix", set.Prefix)
}

func TestExtensionInterceptor_WithoutOption(t *testing.T) {
	Interceptor()

	assert.NotNil(t, optionsMap[rkgininter.RpcEntryNameValue])
}

func TestExtensionInterceptor_HappyCase(t *testing.T) {
	Interceptor(WithEntryNameAndType("ut-name", "ut-type"))

	assert.NotNil(t, optionsMap["ut-name"])
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginbasic

import (
	"github.com/gin-gonic/gin"
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

func TestGetOptionSet_WithNilContext(t *testing.T) {
	assert.Nil(t, GetOptionSet(nil))
}

func TestGetOptionSet_WithEmptyEntryNameInContext(t *testing.T) {
	BasicInterceptor(WithEntryNameAndType("ut-name", "ut-type"))
	assert.Nil(t, GetOptionSet(&gin.Context{}))
}

func TestGetOptionSet_HappyCase(t *testing.T) {
	BasicInterceptor(WithEntryNameAndType("ut-name", "ut-type"))
	ctx := &gin.Context{
		Keys: map[string]interface{}{
			RkEntryNameKey: "ut-name",
		},
	}

	assert.NotNil(t, GetOptionSet(ctx))
}

func TestBasicInterceptor_WithoutOption(t *testing.T) {
	BasicInterceptor()

	assert.NotNil(t, optionsMap[RkEntryNameValue])
}

func TestBasicInterceptor_HappyCase(t *testing.T) {
	BasicInterceptor(WithEntryNameAndType("ut-name", "ut-type"))

	assert.NotNil(t, optionsMap["ut-name"])
}

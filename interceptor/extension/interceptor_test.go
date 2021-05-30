package rkginextension

import (
	"github.com/rookie-ninja/rk-gin/interceptor/context"
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
	ExtensionInterceptor()

	assert.NotNil(t, optionsMap[rkginctx.RkEntryNameValue])
}

func TestExtensionInterceptor_HappyCase(t *testing.T) {
	ExtensionInterceptor(WithEntryNameAndType("ut-name", "ut-type"))

	assert.NotNil(t, optionsMap["ut-name"])
}

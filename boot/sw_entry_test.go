// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithPath_HappyCase(t *testing.T) {
	entry := NewSwEntry(WithPathSw("ut-path"))
	assert.Equal(t, "/ut-path/", entry.Path)
}

func TestWithHeaders_HappyCase(t *testing.T) {
	headers := map[string]string{
		"key": "value",
	}
	entry := NewSwEntry(WithHeadersSw(headers))
	assert.Len(t, entry.Headers, 1)
}

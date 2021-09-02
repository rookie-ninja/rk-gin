// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkginctx

import (
	"github.com/gin-gonic/gin"
	// httptest "github.com/stretchr/testify/http"
	"os"
	"testing"
)

var (
	key   = "key"
	value = "value"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	code := m.Run()
	os.Exit(code)
}

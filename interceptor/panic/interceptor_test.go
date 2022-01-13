// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginpanic

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	ctx, router := gin.CreateTestContext(httptest.NewRecorder())
	router.Use(Interceptor(
		rkmidpanic.WithEntryNameAndType("ut-entry", "ut-type")))
	router.Handle(http.MethodGet, "/ut", func(context *gin.Context) {
		panic(errors.New("ut panic"))
	})

	ctx.Request = httptest.NewRequest(http.MethodGet, "/ut", nil)
	router.HandleContext(ctx)
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

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintimeout

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sleepHandler(ctx *gin.Context) {
	time.Sleep(time.Second)
	ctx.JSON(http.StatusOK, "{}")
}

func panicHandler(ctx *gin.Context) {
	panic(fmt.Errorf("ut panic"))
}

func returnHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "{}")
}

var customResponse = func(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, "{}")
}

func getGinRouter(path string, handler gin.HandlerFunc, middleware gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(middleware)
	r.GET(path, handler)
	return r
}

func TestInterceptor_WithTimeout(t *testing.T) {
	// with global timeout response
	r := getGinRouter("/", sleepHandler, Interceptor(
		WithTimeoutAndResp(time.Nanosecond, nil)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// with path
	r = getGinRouter("/ut-path", sleepHandler, Interceptor(
		WithTimeoutAndRespByPath("/ut-path", time.Nanosecond, nil)))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ut-path", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// with custom response
	r = getGinRouter("/", sleepHandler, Interceptor(
		WithTimeoutAndRespByPath("/", time.Nanosecond, customResponse)))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestInterceptor_WithPanic(t *testing.T) {
	defer assertPanic(t)

	r := getGinRouter("/", panicHandler, Interceptor(
		WithTimeoutAndResp(time.Minute, nil)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
}

func TestInterceptor_HappyCase(t *testing.T) {
	// Let's add two routes /timeout and /happy
	// We expect interceptor acts as the name describes
	r := gin.New()
	r.Use(Interceptor(
		WithTimeoutAndRespByPath("/timeout", time.Nanosecond, nil),
		WithTimeoutAndRespByPath("/happy", time.Minute, nil)))

	r.GET("/timeout", sleepHandler)
	r.GET("/happy", returnHandler)

	// timeout on /timeout
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/timeout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// OK on /happy
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/happy", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgintout

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/middleware/timeout"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func sleepH(ctx *gin.Context) {
	time.Sleep(time.Second)
	ctx.JSON(http.StatusOK, "{}")
}

func panicH(ctx *gin.Context) {
	panic(fmt.Errorf("ut panic"))
}

func returnH(ctx *gin.Context) {
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
	r := getGinRouter("/", sleepH, Middleware(
		rkmidtimeout.WithTimeout(time.Nanosecond)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// with path
	r = getGinRouter("/ut-path", sleepH, Middleware(
		rkmidtimeout.WithTimeoutByPath("/ut-path", time.Nanosecond)))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ut-path", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestInterceptor_WithPanic(t *testing.T) {
	defer assertPanic(t)

	r := getGinRouter("/", panicH, Middleware(
		rkmidtimeout.WithTimeout(time.Minute)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
}

func TestInterceptor_HappyCase(t *testing.T) {
	// Let's add two routes /timeout and /happy
	// We expect interceptor acts as the name describes
	r := gin.New()
	r.Use(Middleware(
		rkmidtimeout.WithTimeoutByPath("/timeout", time.Nanosecond),
		rkmidtimeout.WithTimeoutByPath("/happy", time.Minute)))

	r.GET("/timeout", sleepH)
	r.GET("/happy", returnH)

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

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

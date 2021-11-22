// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgincors

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const originHeaderValue = "http://ut-origin"

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	router := getRouter(Interceptor(WithSkipper(func(context *gin.Context) bool {
		return true
	})))
	resp := performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusOK, resp.Code)

	// with empty option, all request will be passed
	router = getRouter(Interceptor())
	resp = performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusOK, resp.Code)

	// match 1.1
	router = getRouter(Interceptor())
	resp = performRequest(router, http.MethodGet)
	assert.Equal(t, http.StatusOK, resp.Code)

	// match 1.2
	router = getRouter(Interceptor())
	resp = performRequest(router, http.MethodOptions)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	// match 2.1
	router = getRouter(Interceptor(WithAllowOrigins("http://do-not-pass-through")))
	resp = performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusFound, resp.Code)

	// match 2.2
	router = getRouter(Interceptor(WithAllowOrigins("http://do-not-pass-through")))
	resp = performRequest(router, http.MethodOptions, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusNoContent, resp.Code)

	// match 3
	router = getRouter(Interceptor())
	resp = performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))

	// match 3.1
	router = getRouter(Interceptor(WithAllowCredentials(true)))
	resp = performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header().Get(headerAccessControlAllowCredentials))

	// match 3.2
	router = getRouter(Interceptor(
		WithAllowCredentials(true),
		WithExposeHeaders("expose")))
	resp = performRequest(router, http.MethodGet, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header().Get(headerAccessControlAllowCredentials))
	assert.Equal(t, "expose", resp.Header().Get(headerAccessControlExposeHeaders))

	// match 4
	router = getRouter(Interceptor())
	resp = performRequest(router, http.MethodOptions, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header().Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))

	// match 4.1
	router = getRouter(Interceptor(WithAllowCredentials(true)))
	resp = performRequest(router, http.MethodOptions, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header().Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header().Get(headerAccessControlAllowMethods))
	assert.Equal(t, "true", resp.Header().Get(headerAccessControlAllowCredentials))

	// match 4.2
	router = getRouter(Interceptor(WithAllowHeaders("ut-header")))
	resp = performRequest(router, http.MethodOptions, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header().Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header().Get(headerAccessControlAllowMethods))
	assert.Equal(t, "ut-header", resp.Header().Get(headerAccessControlAllowHeaders))

	// match 4.3
	router = getRouter(Interceptor(WithMaxAge(1)))
	resp = performRequest(router, http.MethodOptions, header{headerOrigin, originHeaderValue})
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, []string{
		headerAccessControlRequestMethod,
		headerAccessControlRequestHeaders,
	}, resp.Header().Values(headerVary))
	assert.Equal(t, originHeaderValue, resp.Header().Get(headerAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header().Get(headerAccessControlAllowMethods))
	assert.Equal(t, "1", resp.Header().Get(headerAccessControlMaxAge))
}

func getRouter(middleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middleware...)
	router.GET("/get", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "")
	})

	return router
}

func performRequest(r http.Handler, method string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/get", nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

type header struct {
	Key   string
	Value string
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

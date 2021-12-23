// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginjwt

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	handler := Interceptor(WithSkipper(func(context *gin.Context) bool {
		return true
	}))
	ctx := newCtx()
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())

	// without options
	handler = Interceptor()
	ctx = newCtx()
	handler(ctx)
	assert.Equal(t, http.StatusUnauthorized, ctx.Writer.Status())

	// with parse token error
	parseTokenErrFunc := func(auth string, c *gin.Context) (*jwt.Token, error) {
		return nil, errors.New("ut-error")
	}
	handler = Interceptor(
		WithParseTokenFunc(parseTokenErrFunc))
	ctx = newCtx()
	handler(ctx)
	assert.Equal(t, http.StatusUnauthorized, ctx.Writer.Status())

	// happy case
	parseTokenErrFunc = func(auth string, c *gin.Context) (*jwt.Token, error) {
		return &jwt.Token{}, nil
	}
	handler = Interceptor(
		WithParseTokenFunc(parseTokenErrFunc))
	ctx = newCtx()
	ctx.Request.Header.Set(headerAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	handler(ctx)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
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

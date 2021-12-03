// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgincsrf

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestNewOptionSet(t *testing.T) {
	ctx, _ := gin.CreateTestContext(&httptest.ResponseRecorder{})
	// without options
	set := newOptionSet()
	assert.NotEmpty(t, set.EntryName)
	assert.NotEmpty(t, set.EntryType)
	assert.False(t, set.Skipper(ctx))
	assert.Equal(t, 32, set.TokenLength)
	assert.Equal(t, "header:"+headerXCSRFToken, set.TokenLookup)
	assert.Equal(t, "_csrf", set.CookieName)
	assert.Equal(t, 86400, set.CookieMaxAge)
	assert.Equal(t, http.SameSiteDefaultMode, set.CookieSameSite)
	assert.Empty(t, set.IgnorePrefix)
	assert.NotNil(t, set.extractor)

	// with options
	set = newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithSkipper(func(context *gin.Context) bool {
			return true
		}),
		WithTokenLength(10),
		WithTokenLookup("header:ut-header"),
		WithCookieName("ut-cookie"),
		WithCookieDomain("ut-domain"),
		WithCookiePath("ut-path"),
		WithCookieMaxAge(10),
		WithCookieHTTPOnly(true),
		WithCookieSameSite(http.SameSiteDefaultMode),
	)
	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
	assert.True(t, set.Skipper(ctx))
	assert.Equal(t, 10, set.TokenLength)
	assert.Equal(t, "header:ut-header", set.TokenLookup)
	assert.Equal(t, "ut-cookie", set.CookieName)
	assert.Equal(t, "ut-domain", set.CookieDomain)
	assert.Equal(t, "ut-path", set.CookiePath)
	assert.True(t, set.CookieHTTPOnly)
	assert.Equal(t, 10, set.CookieMaxAge)
	assert.Equal(t, http.SameSiteDefaultMode, set.CookieSameSite)
	assert.Empty(t, set.IgnorePrefix)
	assert.NotNil(t, set.extractor)
}

func TestIsValidToken(t *testing.T) {
	// expect ture
	token := "my-token"
	clientToken := "my-token"

	assert.True(t, isValidToken(token, clientToken))

	// expect false
	assert.False(t, isValidToken(token, clientToken+"-invalid"))
}

func TestCsrfTokenFromHeader(t *testing.T) {
	set := newOptionSet(WithTokenLookup("header:ut-header"))

	// happy case
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("ut-header", "ut-header-value")
	ctx, _ := gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req

	res, err := set.extractor(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "ut-header-value", res)

	// expect error
	req = &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("ut-header-invalid", "ut-header-value")
	ctx, _ = gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req
	res, err = set.extractor(ctx)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestCsrfTokenFromForm(t *testing.T) {
	set := newOptionSet(WithTokenLookup("form:ut-form"))

	// happy case
	req := &http.Request{
		Form: url.Values{},
	}
	req.Form.Set("ut-form", "ut-form-value")
	ctx, _ := gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req

	res, err := set.extractor(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "ut-form-value", res)

	// expect error
	req = &http.Request{
		Form: url.Values{},
	}
	req.Form.Set("ut-form-invalid", "ut-form-value")
	ctx, _ = gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req

	res, err = set.extractor(ctx)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestCsrfTokenFromQuery(t *testing.T) {
	set := newOptionSet(WithTokenLookup("query:ut-query"))

	// happy case
	req := &http.Request{
		URL: &url.URL{},
	}
	req.URL.RawQuery = "ut-query=ut-query-value"
	ctx, _ := gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req

	res, err := set.extractor(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "ut-query-value", res)

	// expect error
	req = &http.Request{
		URL: &url.URL{},
	}
	req.URL.RawQuery = "ut-query-invalid=ut-query-value"
	ctx, _ = gin.CreateTestContext(&httptest.ResponseRecorder{})
	ctx.Request = req

	res, err = set.extractor(ctx)
	assert.NotNil(t, err)
	assert.Empty(t, res)
}

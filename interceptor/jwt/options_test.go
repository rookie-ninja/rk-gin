// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginjwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestNewOptionSet(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)

	// without options
	set := newOptionSet()
	assert.NotEmpty(t, set.EntryName)
	assert.NotEmpty(t, set.EntryType)
	assert.False(t, set.Skipper(ctx))
	assert.Empty(t, set.SigningKeys)
	assert.Nil(t, set.SigningKey)
	assert.Equal(t, set.SigningAlgorithm, AlgorithmHS256)
	assert.NotNil(t, set.Claims)
	assert.Equal(t, set.TokenLookup, "header:"+headerAuthorization)
	assert.Equal(t, set.AuthScheme, "Bearer")
	assert.Equal(t, reflect.ValueOf(set.KeyFunc).Pointer(), reflect.ValueOf(set.defaultKeyFunc).Pointer())
	assert.Equal(t, reflect.ValueOf(set.ParseTokenFunc).Pointer(), reflect.ValueOf(set.defaultParseToken).Pointer())

	// with options
	skipper := func(*gin.Context) bool {
		return false
	}
	claims := &fakeClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	}
	parseToken := func(string, *gin.Context) (*jwt.Token, error) { return nil, nil }
	tokenLookups := strings.Join([]string{
		"query:ut-query",
		"param:ut-param",
		"cookie:ut-cookie",
		"form:ut-form",
		"header:ut-header",
	}, ",")

	set = newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithSkipper(skipper),
		WithSigningKey("ut-signing-key"),
		WithSigningKeys("ut-key", "ut-value"),
		WithSigningAlgorithm("ut-signing-algorithm"),
		WithClaims(claims),
		WithTokenLookup(tokenLookups),
		WithAuthScheme("ut-auth-scheme"),
		WithKeyFunc(keyFunc),
		WithParseTokenFunc(parseToken),
		WithIgnorePrefix("/ut"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
	assert.False(t, set.Skipper(ctx))
	assert.Equal(t, "ut-signing-key", set.SigningKey)
	assert.NotEmpty(t, set.SigningKeys)
	assert.Equal(t, "ut-signing-algorithm", set.SigningAlgorithm)
	assert.Equal(t, claims, set.Claims)
	assert.Equal(t, tokenLookups, set.TokenLookup)
	assert.Len(t, set.extractors, 5)
	assert.Equal(t, "ut-auth-scheme", set.AuthScheme)
	assert.Equal(t, reflect.ValueOf(set.KeyFunc).Pointer(), reflect.ValueOf(keyFunc).Pointer())
	assert.Equal(t, reflect.ValueOf(set.ParseTokenFunc).Pointer(), reflect.ValueOf(parseToken).Pointer())
}

func TestJwtFromHeader(t *testing.T) {
	headerKey := "ut-header"
	authScheme := "ut-auth-scheme"
	jwtValue := "ut-jwt"
	extractor := jwtFromHeader(headerKey, authScheme)
	ctx := newCtx()

	// happy case
	ctx.Request.Header.Set(headerKey, strings.Join([]string{authScheme, jwtValue}, " "))
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx.Request.Header.Set(headerKey, strings.Join([]string{"invalid", jwtValue}, " "))
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromQuery(t *testing.T) {
	queryKey := "ut-query"
	jwtValue := "ut-jwt"
	extractor := jwtFromQuery(queryKey)
	ctx := newCtx()

	// happy case
	ctx.Request.URL.RawQuery = strings.Join([]string{queryKey, jwtValue}, "=")
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx = newCtx()
	ctx.Request.URL.RawQuery = strings.Join([]string{"invalid", jwtValue}, "=")
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromParam(t *testing.T) {
	paramKey := "ut-param"
	jwtValue := "ut-jwt"
	extractor := jwtFromParam(paramKey)
	ctx := newCtx()

	// happy case
	ctx.Params = gin.Params{
		gin.Param{
			Key:   paramKey,
			Value: jwtValue,
		},
	}
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx = newCtx()
	ctx.Params = gin.Params{
		gin.Param{
			Key:   "invalid",
			Value: jwtValue,
		},
	}
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromCookie(t *testing.T) {
	cookieKey := "ut-cookie"
	jwtValue := "ut-jwt"
	extractor := jwtFromCookie(cookieKey)
	ctx := newCtx()

	// happy case
	ctx.Request.AddCookie(&http.Cookie{
		Name:  cookieKey,
		Value: jwtValue,
	})
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx = newCtx()
	ctx.Request.AddCookie(&http.Cookie{
		Name:  "invalid",
		Value: jwtValue,
	})
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromForm(t *testing.T) {
	formKey := "ut-form"
	jwtValue := "ut-jwt"
	extractor := jwtFromForm(formKey)
	ctx := newCtx()

	// happy case
	ctx.Request.Form = url.Values{
		formKey: []string{jwtValue},
	}
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx = newCtx()
	ctx.Request.Form = url.Values{
		"invalid": []string{jwtValue},
	}
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func newCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(NewMockResponseWriter())
	ctx.Request = &http.Request{
		Header: make(map[string][]string, 0),
		URL:    &url.URL{},
	}

	return ctx
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		data:   make([]byte, 0),
		header: http.Header{},
	}
}

type MockResponseWriter struct {
	data       []byte
	statusCode int
	header     http.Header
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(bytes []byte) (int, error) {
	m.data = bytes
	return len(bytes), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

type fakeClaims struct{}

func (c *fakeClaims) Valid() error {
	return nil
}

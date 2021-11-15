// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgingzip

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getBody(encode bool) io.Reader {
	if encode {
		buf := new(bytes.Buffer)
		zw := gzip.NewWriter(buf)
		zw.Write([]byte("ut-string"))
		zw.Flush()
		zw.Close()
		return io.NopCloser(buf)
	}

	buf := new(bytes.Buffer)
	buf.WriteString("ut-string")
	return io.NopCloser(buf)
}

func readResponse(encode bool, r io.Reader) string {
	if encode {
		zr, _ := gzip.NewReader(r)
		buf := new(bytes.Buffer)
		io.Copy(buf, zr)
		return buf.String()
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	return buf.String()
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// 1: without skipper
	router := gin.New()
	router.Use(Interceptor())
	router.POST("/post", func(ctx *gin.Context) {
		buf := new(strings.Builder)
		io.Copy(buf, ctx.Request.Body)

		ctx.String(http.StatusOK, buf.String())
	})
	resp := performRequest(router, http.MethodPost, "/post", getBody(true),
		header{headerContentEncoding, gzipEncoding},
		header{headerAcceptEncoding, gzipEncoding})
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "ut-string", readResponse(true, resp.Body))

	// 2: with skipper
	router = gin.New()
	router.Use(Interceptor(WithSkipper(func(context *gin.Context) bool {
		return true
	})))
	router.POST("/post", func(ctx *gin.Context) {
		buf := new(strings.Builder)
		io.Copy(buf, ctx.Request.Body)
		ctx.String(http.StatusOK, buf.String())
	})
	resp = performRequest(router, http.MethodPost, "/post", getBody(true),
		header{headerContentEncoding, gzipEncoding},
		header{headerAcceptEncoding, gzipEncoding})
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "ut-string", readResponse(true, resp.Body))

	// 3: with invalid gzip content
	router = gin.New()
	router.Use(Interceptor())
	router.POST("/post", func(ctx *gin.Context) {
		buf := new(strings.Builder)
		io.Copy(buf, ctx.Request.Body)
		ctx.String(http.StatusOK, buf.String())
	})
	resp = performRequest(router, http.MethodPost, "/post", getBody(false),
		header{headerContentEncoding, gzipEncoding},
		header{headerAcceptEncoding, gzipEncoding})
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	// 4: with empty response body
	router = gin.New()
	router.Use(Interceptor())
	router.POST("/post", func(ctx *gin.Context) {})
	resp = performRequest(router, http.MethodPost, "/post", getBody(true),
		header{headerContentEncoding, gzipEncoding},
		header{headerAcceptEncoding, gzipEncoding})
	assert.Equal(t, http.StatusOK, resp.Code)
}

func performRequest(r http.Handler, method, path string, body io.Reader, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
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

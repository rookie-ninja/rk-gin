// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgingzip

import (
	"bytes"
	"github.com/gin-gonic/gin"
	rkerror "github.com/rookie-ninja/rk-entry/error"
	"github.com/rookie-ninja/rk-entry/middleware"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Middleware Add gzip compress and decompress interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/decompress.go
// https://github.com/labstack/echo/blob/master/middleware/compress.go
func Middleware(opts ...Option) gin.HandlerFunc {
	set := newOptionSet(opts...)

	return func(ctx *gin.Context) {
		ctx.Set(rkmid.EntryNameKey.String(), set.EntryName)

		if set.Skipper(ctx) || set.ignore(ctx) {
			ctx.Next()
			return
		}

		// deal with request decompression
		switch ctx.Request.Header.Get(headerContentEncoding) {
		case gzipEncoding:
			gzipReader := set.decompressPool.Get()

			// make gzipReader to read from original request body
			if err := gzipReader.Reset(ctx.Request.Body); err != nil {
				// return reader back to sync.Pool
				set.decompressPool.Put(gzipReader)

				// body is empty, keep on going
				if err == io.EOF {
					ctx.Next()
					return
				}

				ctx.AbortWithStatusJSON(http.StatusInternalServerError, rkerror.NewInternalError(err))

				return
			}

			// create a buffer and copy decompressed data into it via gzipReader
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, gzipReader); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, rkerror.NewInternalError(err))
				return
			}

			// close both gzipReader and original reader in request body
			gzipReader.Close()
			ctx.Request.Body.Close()
			set.decompressPool.Put(gzipReader)

			// assign decompressed buffer to request
			ctx.Request.Body = ioutil.NopCloser(&buf)
		}

		// deal with response compression
		ctx.Writer.Header().Add(headerVary, headerAcceptEncoding)
		// gzip is one of expected encoding type from request
		if strings.Contains(ctx.Request.Header.Get(headerAcceptEncoding), gzipEncoding) {
			// set to response header
			ctx.Writer.Header().Set(headerContentEncoding, gzipEncoding)

			// create gzip writer
			gzipWriter := set.compressPool.Get()

			// reset writer of gzip writer to original writer from response
			originalWriter := ctx.Writer
			gzipWriter.Reset(originalWriter)

			// defer func
			defer func() {
				if ctx.Writer.Size() == -1 {
					// remove encoding header if response is empty
					if ctx.Writer.Header().Get(headerContentEncoding) == gzipEncoding {
						ctx.Writer.Header().Del(headerContentEncoding)
					}
					// we have to reset response to it's pristine state when
					// nothing is written to body or error is returned.
					ctx.Writer = originalWriter

					// reset to empty
					gzipWriter.Reset(ioutil.Discard)
				}

				// close gzipWriter
				gzipWriter.Close()

				// put gzipWriter back to pool
				set.compressPool.Put(gzipWriter)
			}()

			// assign new writer to response
			ctx.Writer = newGzipResponseWriter(gzipWriter, originalWriter)
		}

		ctx.Next()
	}
}

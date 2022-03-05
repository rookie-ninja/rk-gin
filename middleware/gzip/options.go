// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgingzip

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rs/xid"
	"io/ioutil"
	"strings"
	"sync"
)

const (
	// GzipEncoding encoding type of gzip
	gzipEncoding = "gzip"
	// NoCompression copied from gzip.NoCompression
	NoCompression = "noCompression"
	// BestSpeed copied from gzip.BestSpeed
	BestSpeed = "bestSpeed"
	// BestCompression copied from gzip.BestCompression
	BestCompression = "bestCompression"
	// DefaultCompression copied from gzip.DefaultCompression
	DefaultCompression = "defaultCompression"
	// HuffmanOnly copied from gzip.HuffmanOnly
	HuffmanOnly           = "huffmanOnly"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerAcceptEncoding  = "Accept-Encoding"
)

// Interceptor would distinguish auth set based on.
var (
	optionsMap     = make(map[string]*optionSet)
	defaultSkipper = func(*gin.Context) bool {
		return false
	}
)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:      xid.New().String(),
		EntryType:      "",
		Skipper:        defaultSkipper,
		Level:          DefaultCompression,
		decompressPool: newDecompressPool(),
		ignorePrefix:   make([]string, 0),
	}

	for i := range opts {
		opts[i](set)
	}

	// create a new compressPool
	set.compressPool = newCompressPool(set.Level)

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	EntryName      string
	EntryType      string
	Skipper        Skipper
	Level          string
	decompressPool *decompressPool
	compressPool   *compressPool
	ignorePrefix   []string
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(ctx *gin.Context) bool {
	if ctx.Request != nil && ctx.Request.URL != nil {
		for i := range set.ignorePrefix {
			if strings.HasPrefix(ctx.Request.URL.Path, set.ignorePrefix[i]) {
				return true
			}
		}

		return rkmid.ShouldIgnoreGlobal(ctx.Request.URL.Path)
	}

	return false
}

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// WithLevel provide level of compressing.
func WithLevel(level string) Option {
	return func(opt *optionSet) {
		opt.Level = level
	}
}

// WithSkipper provide skipper.
func WithSkipper(skip Skipper) Option {
	return func(opt *optionSet) {
		opt.Skipper = skip
	}
}

// WithPathToIgnore provide path prefix to ignore middleware
func WithPathToIgnore(prefix ...string) Option {
	return func(opt *optionSet) {
		opt.ignorePrefix = append(opt.ignorePrefix, prefix...)
	}
}

// sync.Pool is the delegate of this pool
type compressPool struct {
	delegate *sync.Pool
}

// Create a new compress pool
func newCompressPool(level string) *compressPool {
	levelLowerCase := strings.ToLower(level)

	levelInt := gzip.DefaultCompression

	switch levelLowerCase {
	case strings.ToLower(NoCompression):
		levelInt = gzip.NoCompression
	case strings.ToLower(BestSpeed):
		levelInt = gzip.BestSpeed
	case strings.ToLower(BestCompression):
		levelInt = gzip.BestCompression
	case strings.ToLower(DefaultCompression):
		levelInt = gzip.DefaultCompression
	case strings.ToLower(HuffmanOnly):
		levelInt = gzip.HuffmanOnly
	default:
		levelInt = gzip.DefaultCompression
	}

	return &compressPool{
		delegate: &sync.Pool{
			New: func() interface{} {
				// Ok to ignore error because of above switch statement
				writer, _ := gzip.NewWriterLevel(ioutil.Discard, levelInt)
				return writer
			},
		},
	}
}

// Get item gzip.Writer from pool
func (p *compressPool) Get() *gzip.Writer {
	// assert no error
	raw := p.delegate.Get()

	switch raw.(type) {
	case *gzip.Writer:
		return raw.(*gzip.Writer)
	}

	return nil
}

// Put item gzip.Writer back to pool
func (p *compressPool) Put(x interface{}) {
	p.delegate.Put(x)
}

// sync.Pool is the delegate of this pool
type decompressPool struct {
	delegate *sync.Pool
}

// Create a new decompress pool
func newDecompressPool() *decompressPool {
	pool := &sync.Pool{
		New: func() interface{} {
			// In order to create a gzip.Reader, we need to pass a bytes with format gzip.
			// Create a gzip.Writer is the easiest way to achieve this goal.
			writer, _ := gzip.NewWriterLevel(ioutil.Discard, gzip.DefaultCompression)
			b := new(bytes.Buffer)
			writer.Reset(b)
			writer.Flush()
			writer.Close()

			// Create a reader, ignoring error since we created a empty writer
			reader, _ := gzip.NewReader(bytes.NewReader(b.Bytes()))
			return reader
		},
	}

	return &decompressPool{
		delegate: pool,
	}
}

// Get item gzip.Reader from pool
func (p *decompressPool) Get() *gzip.Reader {
	// assert no error
	raw := p.delegate.Get()

	switch raw.(type) {
	case *gzip.Reader:
		return raw.(*gzip.Reader)
	}

	return nil
}

// Put item gzip.Reader back to pool
func (p *decompressPool) Put(x interface{}) {
	p.delegate.Put(x)
}

// Skipper default skipper will always return false
type Skipper func(*gin.Context) bool

// Copied from https://github.com/labstack/echo/blob/master/middleware/compress.go
//
// Why not use middleware.GzipWithConfig directly?
//
// rk-echo support multi-entries of echo framework. In order to match rk-echo architecture,
// we need to modify some of logic in middleware.
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func newGzipResponseWriter(w *gzip.Writer, rw gin.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		writer:         w,
		ResponseWriter: rw,
	}
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	g.ResponseWriter.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	g.ResponseWriter.Header().Del("Content-Length")
	return g.writer.Write(data)
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipResponseWriter) WriteHeader(code int) {
	g.ResponseWriter.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

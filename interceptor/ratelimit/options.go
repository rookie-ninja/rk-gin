// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginlimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	juju "github.com/juju/ratelimit"
	"github.com/rookie-ninja/rk-gin/interceptor"
	uber "go.uber.org/ratelimit"
	"strings"
	"time"
)

const (
	TokenBucket   = "tokenBucket"
	LeakyBucket   = "leakyBucket"
	DefaultLimit  = 1000000
	GlobalLimiter = "rk-limiter"
)

// User could implement
type Limiter func(ctx *gin.Context) error

// NoopLimiter will do nothing
type NoopLimiter struct{}

// Limit will do nothing
func (l *NoopLimiter) Limit(*gin.Context) error {
	return nil
}

// ZeroRateLimiter will block requests.
type ZeroRateLimiter struct{}

// Limit will block request and return error
func (l *ZeroRateLimiter) Limit(*gin.Context) error {
	return fmt.Errorf("slow down your request")
}

// tokenBucketLimiter delegates limit logic to juju.Bucket
type tokenBucketLimiter struct {
	delegator *juju.Bucket
}

// Limit delegates limit logic to juju.Bucket
func (l *tokenBucketLimiter) Limit(ctx *gin.Context) error {
	l.delegator.Wait(1)
	return nil
}

// leakyBucketLimiter delegates limit logic to uber.Limiter
type leakyBucketLimiter struct {
	delegator uber.Limiter
}

// Limit delegates limit logic to uber.Limiter
func (l *leakyBucketLimiter) Limit(ctx *gin.Context) error {
	l.delegator.Take()
	return nil
}

// Interceptor would distinguish auth set based on.
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:       rkgininter.RpcEntryNameValue,
		EntryType:       rkgininter.RpcEntryTypeValue,
		reqPerSec:       DefaultLimit,
		reqPerSecByPath: make(map[string]int, DefaultLimit),
		algorithm:       TokenBucket,
		limiter:         make(map[string]Limiter),
	}

	for i := range opts {
		opts[i](set)
	}

	switch set.algorithm {
	case TokenBucket:
		if set.reqPerSec < 1 {
			l := &ZeroRateLimiter{}
			set.setLimiter(GlobalLimiter, l.Limit)
		} else {
			l := &tokenBucketLimiter{
				delegator: juju.NewBucketWithRate(float64(set.reqPerSec), int64(set.reqPerSec)),
			}
			set.setLimiter(GlobalLimiter, l.Limit)
		}

		for k, v := range set.reqPerSecByPath {
			if v < 1 {
				l := &ZeroRateLimiter{}
				set.setLimiter(k, l.Limit)
			} else {
				l := &tokenBucketLimiter{
					delegator: juju.NewBucketWithRate(float64(v), int64(v)),
				}
				set.setLimiter(k, l.Limit)
			}
		}
	case LeakyBucket:
		if set.reqPerSec < 1 {
			l := &ZeroRateLimiter{}
			set.setLimiter(GlobalLimiter, l.Limit)
		} else {
			l := &leakyBucketLimiter{
				delegator: uber.New(set.reqPerSec),
			}
			set.setLimiter(GlobalLimiter, l.Limit)
		}

		for k, v := range set.reqPerSecByPath {
			if v < 1 {
				l := &ZeroRateLimiter{}
				set.setLimiter(k, l.Limit)
			} else {
				l := &leakyBucketLimiter{
					delegator: uber.New(v),
				}
				set.setLimiter(k, l.Limit)
			}
		}
	default:
		l := &NoopLimiter{}
		set.setLimiter(GlobalLimiter, l.Limit)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Wait until rate limit pass through
func (set *optionSet) Wait(ctx *gin.Context, method string) (time.Duration, error) {
	now := time.Now()

	limiter := set.getLimiter(method)
	if err := limiter(ctx); err != nil {
		return now.Sub(now), err
	}

	return now.Sub(time.Now()), nil
}

func (set *optionSet) getLimiter(method string) Limiter {
	if v, ok := set.limiter[method]; ok {
		return v
	}

	return set.limiter[GlobalLimiter]
}

// Set limiter if not exists
func (set *optionSet) setLimiter(method string, l Limiter) {
	if _, ok := set.limiter[method]; ok {
		return
	}

	set.limiter[method] = l
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	EntryName       string
	EntryType       string
	reqPerSec       int
	reqPerSecByPath map[string]int
	algorithm       string
	limiter         map[string]Limiter
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

// WithReqPerSec Provide request per second.
func WithReqPerSec(reqPerSec int) Option {
	return func(opt *optionSet) {
		if reqPerSec >= 0 {
			opt.reqPerSec = reqPerSec
		}
	}
}

// WithReqPerSecByPath Provide request per second by method.
func WithReqPerSecByPath(path string, reqPerSec int) Option {
	return func(opt *optionSet) {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if reqPerSec >= 0 {
			opt.reqPerSecByPath[path] = reqPerSec
		}
	}
}

// WithAlgorithm provide algorithm of rate limit.
// - tokenBucket
// - leakyBucket
func WithAlgorithm(algo string) Option {
	return func(opt *optionSet) {
		opt.algorithm = algo
	}
}

// WithGlobalLimiter provide user defined Limiter.
func WithGlobalLimiter(l Limiter) Option {
	return func(opt *optionSet) {
		opt.limiter[GlobalLimiter] = l
	}
}

// WithLimiterByPath provide user defined Limiter by method.
func WithLimiterByPath(path string, l Limiter) Option {
	return func(opt *optionSet) {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		opt.limiter[path] = l
	}
}

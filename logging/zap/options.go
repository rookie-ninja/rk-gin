// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_inter_logging

var (
	DefaultOptions = &Options{
		enableLogging: EnableLogging,
		enableMetrics: EnableMetrics,
	}
)

func MergeOpt(opts []Option) *Options {
	optCopy := &Options{}
	*optCopy = *DefaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// Default options
func DisableLogging() bool {
	return false
}

func EnableLogging() bool {
	return true
}

func EnableMetrics() bool {
	return true
}

func DisableMetrics() bool {
	return false
}

type Options struct {
	enableMetrics Enable
	enableLogging Enable
}

type Option func(*Options)

// Implement this if want to enable any functionality among interceptor
type Enable func() bool

func EnableLoggingOption(f Enable) Option {
	return func(o *Options) {
		o.enableLogging = f
	}
}

func EnableMetricsOption(f Enable) Option {
	return func(o *Options) {
		o.enableMetrics = f
	}
}

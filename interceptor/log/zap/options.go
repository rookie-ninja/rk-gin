// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_log

import (
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
)

var (
	defaultOptions = &options{
		enableLogging: true,
		enableMetrics: true,
		eventFactory:  rk_query.NewEventFactory(),
		logger:        rk_logger.NoopLogger,
	}
)

func mergeOpt(opts []Option) {
	for i := range opts {
		opts[i](defaultOptions)
	}
}

type options struct {
	enableMetrics bool
	enableLogging bool
	eventFactory  *rk_query.EventFactory
	logger        *zap.Logger
}

type Option func(*options)

func WithEventFactory(factory *rk_query.EventFactory) Option {
	return func(opt *options) {
		if factory == nil {
			factory = rk_query.NewEventFactory()
		}
		opt.eventFactory = factory
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opt *options) {
		if logger == nil {
			logger = rk_logger.NoopLogger
		}
		opt.logger = logger
	}
}

func WithEnableLogging(enable bool) Option {
	return func(opt *options) {
		opt.enableLogging = enable
	}
}

func WithEnableMetrics(enable bool) Option {
	return func(opt *options) {
		opt.enableMetrics = enable
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"testing"
)

func TestNewTvEntry(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	assert.Equal(t, TvEntryNameDefault, entry.GetName())
	assert.Equal(t, TvEntryType, entry.GetType())
	assert.Equal(t, TvEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestTvEntry_Bootstrap(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	entry.Bootstrap(context.TODO())
}

func TestTvEntry_Interrupt(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	entry.Interrupt(context.TODO())
}

func TestTvEntry_TV(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))
	entry.Bootstrap(context.TODO())

	defer assertNotPanic(t)
	// With nil context
	entry.TV(nil)

	// With all paths
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/apis",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/entries",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/configs",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/certs",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/os",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/env",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/prometheus",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/logs",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/deps",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/license",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/info",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/git",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)

	ctx.Params = []gin.Param{
		{
			Key:   "item",
			Value: "/unknown",
		},
	}
	entry.TV(ctx)
	assert.NotEmpty(t, writer.Output)
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

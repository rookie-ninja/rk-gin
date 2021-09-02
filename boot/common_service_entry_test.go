// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func TestWithNameCommonService_WithEmptyString(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService(""))

	assert.NotEmpty(t, entry.GetName())
}

func TestWithNameCommonService_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("unit-test"))

	assert.Equal(t, "unit-test", entry.GetName())
}

func TestWithEventLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCommonService_HappyCase(t *testing.T) {
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(eventLoggerEntry))

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
}

func TestNewCommonServiceEntry_WithoutOptions(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestNewCommonServiceEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry),
		WithEventLoggerEntryCommonService(eventLoggerEntry),
		WithNameCommonService("ut"))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, "ut", entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestCommonServiceEntry_Bootstrap_WithoutRouter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Bootstrap_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.Background())
}

func TestCommonServiceEntry_Interrupt_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Interrupt(context.Background())
}

func TestCommonServiceEntry_GetName_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("ut"))

	assert.Equal(t, "ut", entry.GetName())
}

func TestCommonServiceEntry_GetType_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.Equal(t, "GinCommonServiceEntry", entry.GetType())
}

func TestCommonServiceEntry_String_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotEmpty(t, entry.String())
}

func TestCommonServiceEntry_Healthy_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry()
	entry.Healthy(nil)
}

func TestCommonServiceEntry_Healthy_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Healthy(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.Equal(t, `{"healthy":true}`, writer.Output)
}

func TestCommonServiceEntry_GC_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Gc(nil)
}

func TestCommonServiceEntry_GC_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Gc(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
}

func TestCommonServiceEntry_Info_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Info(nil)
}

func TestCommonServiceEntry_Info_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Info(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
}

func TestCommonServiceEntry_Config_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Configs(nil)
}

func TestCommonServiceEntry_Config_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	vp := viper.New()
	vp.Set("unit-test-key", "unit-test-value")

	viperEntry := rkentry.RegisterConfigEntry(
		rkentry.WithNameConfig("unit-test"),
		rkentry.WithViperInstanceConfig(vp))

	rkentry.GlobalAppCtx.AddConfigEntry(viperEntry)

	entry.Configs(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
	assert.Contains(t, writer.Output, "unit-test-key")
	assert.Contains(t, writer.Output, "unit-test-value")
}

func TestCommonServiceEntry_APIs_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Apis(nil)
}

func TestCommonServiceEntry_APIs_WithEmptyEntries(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Apis(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
}

func TestCommonServiceEntry_APIs_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	ginEntry := RegisterGinEntry(
		WithCommonServiceEntryGin(entry),
		WithNameGin("unit-test-gin"))
	rkentry.GlobalAppCtx.AddEntry(ginEntry)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Apis(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
}

func TestCommonServiceEntry_Sys_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Sys(nil)
}

func TestCommonServiceEntry_Sys_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Sys(ctx)
	assert.Equal(t, 200, writer.StatusCode)
	assert.NotEmpty(t, writer.Output)
}

func TestCommonServiceEntry_Req_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Req(nil)
}

func TestConstructSwUrl_WithNilEntry(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)
	assert.Equal(t, "N/A", constructSwUrl(nil, ctx))
}

func TestConstructSwUrl_WithNilContext(t *testing.T) {
	path := "ut-sw"
	port := 1111
	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGinEntry(WithSwEntryGin(sw), WithPortGin(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, nil))
}

func TestConstructSwUrl_WithNilRequest(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	path := "ut-sw"
	port := 1111

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGinEntry(WithSwEntryGin(sw), WithPortGin(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_WithEmptyHost(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Request = &http.Request{
		Host: "",
	}

	path := "ut-sw"
	port := 1111

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterGinEntry(WithSwEntryGin(sw), WithPortGin(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Request = &http.Request{
		Host: "8.8.8.8:1111",
	}

	path := "ut-sw"
	port := 1111

	sw := NewSwEntry(WithPathSw(path), WithPortSw(uint64(port)))
	entry := RegisterGinEntry(WithSwEntryGin(sw), WithPortGin(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestContainsMetrics_ExpectFalse(t *testing.T) {
	api := "/rk/v1/non-exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: "/rk/v1/exist",
	})

	assert.False(t, containsMetrics(api, metrics))
}

func TestContainsMetrics_ExpectTrue(t *testing.T) {
	api := "/rk/v1/exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: api,
	})

	assert.True(t, containsMetrics(api, metrics))
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	code := m.Run()
	os.Exit(code)
}

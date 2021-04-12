package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestLoggingInterceptor_HappyCase(t *testing.T) {
	handler := LoggingZapInterceptor()

	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})
	ctx.Request = &http.Request{
		Method: "unit-test-method",
		Proto:  "unit-test-proto",
		URL: &url.URL{
			Path: "unit-test-path",
		},
	}

	// call interceptor
	handler(ctx)

	// 1: event should be added into context
	value, exist := ctx.Get(rkginctx.RKEventKey)
	event := value.(rkquery.Event)
	assert.True(t, exist)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.GetRemoteAddr())
	assert.NotEmpty(t, event.GetOperation())
	assert.NotEmpty(t, event.GetFields())
	assert.Empty(t, event.GetEventId())
	assert.Equal(t, "Ended", event.GetEventStatus().String())

	// 2: logger should be added into context with incoming request ids
	value, exist = ctx.Get(rkginctx.RKLoggerKey)
	logger := value.(*zap.Logger)
	assert.True(t, exist)
	assert.NotNil(t, logger)
}

const unknown = "unknown"

func TestDefaultVariables_HappyCase(t *testing.T) {
	assert.Equal(t, unknown, rkginctx.Realm.String)
	assert.Equal(t, unknown, rkginctx.Region.String)
	assert.Equal(t, unknown, rkginctx.AZ.String)
	assert.Equal(t, unknown, rkginctx.Domain.String)
	assert.Equal(t, unknown, rkginctx.AppVersion.String)
	assert.Equal(t, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, rkentry.AppNameDefault)
	assert.NotEmpty(t, rkginctx.LocalIP.String)
	assert.NotEmpty(t, rkginctx.LocalHostname.String)
}

func TestGetEnvValueOrDefault_ExpectEnvValue(t *testing.T) {
	assert.Nil(t, os.Setenv("key", "value"))
	assert.Equal(t, "value", rkcommon.GetEnvValueOrDefault("key", "default"))
	assert.Nil(t, os.Unsetenv("key"))
}

func TestGetEnvValueOrDefault_ExpectDefaultValue(t *testing.T) {
	assert.Equal(t, "default", rkcommon.GetEnvValueOrDefault("key", "default"))
}

func TestGetRemoteAddressSet_WithNilContext(t *testing.T) {
	set := rkginctx.GetRemoteAddressSet(nil)
	assert.Len(t, set, 2)
	assert.Equal(t, "0.0.0.0", set[0].String)
	assert.Equal(t, "0", set[1].String)
}

func TestGetRemoteAddressSet_WithNilRequest(t *testing.T) {
	set := rkginctx.GetRemoteAddressSet(&gin.Context{})
	assert.Len(t, set, 2)
	assert.Equal(t, "0.0.0.0", set[0].String)
	assert.Equal(t, "0", set[1].String)
}

func TestGetRemoteAddressSet_WithInvalidRemoteAddr(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			RemoteAddr: "",
		},
	}

	set := rkginctx.GetRemoteAddressSet(ctx)
	assert.Len(t, set, 2)
	assert.Equal(t, "0.0.0.0", set[0].String)
	assert.Equal(t, "0", set[1].String)
}

func TestGetRemoteAddressSet_HappyCase(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			RemoteAddr: "localhost:1949",
		},
	}

	set := rkginctx.GetRemoteAddressSet(ctx)
	assert.NotEmpty(t, set)
	assert.Equal(t, "localhost", set[0].String)
	assert.Equal(t, "1949", set[1].String)
}

func TestGetRemoteAddressSet_WithForwardedIP(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			RemoteAddr: "localhost:1949",
			Header:     make(map[string][]string),
		},
	}

	ctx.Request.Header.Set("x-forwarded-for", "1.1.1.1")
	set := rkginctx.GetRemoteAddressSet(ctx)
	assert.NotEmpty(t, set)
	assert.Equal(t, "1.1.1.1", set[0].String)
	assert.Equal(t, "1949", set[1].String)
}

func TestGetRemoteAddressSet_WithForwardedSpecialIP(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			RemoteAddr: "localhost:1949",
			Header:     make(map[string][]string),
		},
	}

	ctx.Request.Header.Set("x-forwarded-for", "::1")

	set := rkginctx.GetRemoteAddressSet(ctx)
	assert.NotEmpty(t, set)
	assert.Equal(t, "localhost", set[0].String)
	assert.Equal(t, "1949", set[1].String)
}

func TestWithEventFactory_WithNilFactory(t *testing.T) {
	Opt := WithEventFactory(nil)
	defaultOptions := &options{
		eventFactory: rkquery.NewEventFactory(),
		logger:       rklogger.StdoutLogger,
	}
	Opt(defaultOptions)

	assert.NotNil(t, defaultOptions.eventFactory)
}

func TestWithEventFactory_HappyCase(t *testing.T) {
	factory := rkquery.NewEventFactory()
	Opt := WithEventFactory(factory)
	defaultOptions := &options{
		eventFactory: rkquery.NewEventFactory(),
		logger:       rklogger.StdoutLogger,
	}
	Opt(defaultOptions)

	assert.Equal(t, factory, defaultOptions.eventFactory)
}

func TestWithLogger_WithNilFactory(t *testing.T) {
	Opt := WithLogger(nil)
	defaultOptions := &options{
		eventFactory: rkquery.NewEventFactory(),
		logger:       rklogger.StdoutLogger,
	}
	Opt(defaultOptions)

	assert.NotNil(t, defaultOptions.logger)
}

func TestWithLogger_HappyCase(t *testing.T) {
	logger := rklogger.NoopLogger
	Opt := WithLogger(logger)
	defaultOptions := &options{
		eventFactory: rkquery.NewEventFactory(),
		logger:       rklogger.StdoutLogger,
	}
	Opt(defaultOptions)

	assert.Equal(t, logger, defaultOptions.logger)
}

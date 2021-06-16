package rkginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/common"
	rkginbasic "github.com/rookie-ninja/rk-gin/interceptor/basic"
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

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestLoggingInterceptor_HappyCase(t *testing.T) {
	handler := LoggingZapInterceptor(
		WithEventFactory(
			rkquery.NewEventFactory(
				rkquery.WithZapLogger(rklogger.NoopLogger))))

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
	value, exist := ctx.Get(rkginbasic.RkEventKey)
	event := value.(rkquery.Event)
	assert.True(t, exist)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.GetRemoteAddr())
	assert.NotEmpty(t, event.GetOperation())
	assert.NotEmpty(t, event.ListPayloads())
	assert.NotEmpty(t, event.GetEventId())
	assert.Equal(t, "Ended", event.GetEventStatus().String())

	// 2: logger should be added into context with incoming request ids
	value, exist = ctx.Get(rkginbasic.RkLoggerKey)
	logger := value.(*zap.Logger)
	assert.True(t, exist)
	assert.NotNil(t, logger)
}

func TestDefaultVariables_HappyCase(t *testing.T) {
	assert.Equal(t, "*", rkginbasic.Realm.String)
	assert.Equal(t, "*", rkginbasic.Region.String)
	assert.Equal(t, "*", rkginbasic.AZ.String)
	assert.Equal(t, "*", rkginbasic.Domain.String)
	assert.NotEmpty(t, rkginbasic.LocalIp.String)
	assert.NotEmpty(t, rkginbasic.LocalHostname.String)
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
	set := getRemoteAddressSet(nil)
	assert.Len(t, set, 2)
	assert.Equal(t, "0.0.0.0", set[0].String)
	assert.Equal(t, "0", set[1].String)
}

func TestGetRemoteAddressSet_WithNilRequest(t *testing.T) {
	set := getRemoteAddressSet(&gin.Context{})
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

	set := getRemoteAddressSet(ctx)
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

	set := getRemoteAddressSet(ctx)
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
	set := getRemoteAddressSet(ctx)
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

	set := getRemoteAddressSet(ctx)
	assert.NotEmpty(t, set)
	assert.Equal(t, "localhost", set[0].String)
	assert.Equal(t, "1949", set[1].String)
}

func TestWithEventFactory_WithNilFactory(t *testing.T) {
	Opt := WithEventFactory(nil)
	set := &optionSet{
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
	}
	Opt(set)

	assert.NotNil(t, set.EventFactory)
}

func TestWithEventFactory_HappyCase(t *testing.T) {
	factory := rkquery.NewEventFactory()
	Opt := WithEventFactory(factory)
	set := &optionSet{
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
	}
	Opt(set)

	assert.Equal(t, factory, set.EventFactory)
}

func TestWithLogger_WithNilFactory(t *testing.T) {
	Opt := WithLogger(nil)
	set := &optionSet{
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
	}
	Opt(set)

	assert.NotNil(t, set.Logger)
}

func TestWithLogger_HappyCase(t *testing.T) {
	logger := rklogger.NoopLogger
	Opt := WithLogger(logger)
	set := &optionSet{
		EventFactory: rkquery.NewEventFactory(),
		Logger:       rklogger.StdoutLogger,
	}
	Opt(set)

	assert.Equal(t, logger, set.Logger)
}

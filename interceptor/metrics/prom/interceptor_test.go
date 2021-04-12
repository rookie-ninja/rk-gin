package rkginmetrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	rkprom "github.com/rookie-ninja/rk-prom"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestMetricsInterceptor_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	handler := MetricsPromInterceptor()
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})

	// call interceptor
	handler(ctx)

	// unregister metrics
	clearAllMetrics()
}

func TestDefaultMetricsVariables_HappyCase(t *testing.T) {
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerMetricsSet(rkginctx.RKEntryDefaultName))
	assert.NotEmpty(t, GetServerMetricsSet(rkginctx.RKEntryDefaultName))

	// server metrics
	assert.Equal(t, "rkapp", GetServerMetricsSet(rkginctx.RKEntryDefaultName).GetNamespace())
	assert.Equal(t, "entry", GetServerMetricsSet(rkginctx.RKEntryDefaultName).GetSubSystem())

	// default labels
	assert.Contains(t, DefaultLabelKeys, "realm")
	assert.Contains(t, DefaultLabelKeys, "region")
	assert.Contains(t, DefaultLabelKeys, "az")
	assert.Contains(t, DefaultLabelKeys, "domain")
	assert.Contains(t, DefaultLabelKeys, "app_version")
	assert.Contains(t, DefaultLabelKeys, "app_name")
	assert.Contains(t, DefaultLabelKeys, "method")
	assert.Contains(t, DefaultLabelKeys, "path")
	assert.Contains(t, DefaultLabelKeys, "res_code")

	// unregister metrics
	clearAllMetrics()
}

func TestInitMetrics_HappyCase(t *testing.T) {
	defaultOptions := &options{
		entryName: rkginctx.RKEntryDefaultName,
		metricsSet: rkprom.NewMetricsSet(
			rkentry.GlobalAppCtx.GetAppInfoEntry().AppName,
			rkginctx.RKEntryDefaultName,
			prometheus.DefaultRegisterer),
	}

	initMetrics(defaultOptions)

	// metrics
	assert.Equal(t, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, defaultOptions.metricsSet.GetNamespace())
	assert.Equal(t, defaultOptions.entryName, defaultOptions.metricsSet.GetSubSystem())
	assert.NotNil(t, defaultOptions.metricsSet.GetCounter(Errors))
	assert.NotNil(t, defaultOptions.metricsSet.GetCounter(ResCode))
	assert.NotNil(t, defaultOptions.metricsSet.GetSummary(ElapsedNano))

	// unregister metrics
	defaultOptions.metricsSet.UnRegisterCounter(Errors)
	defaultOptions.metricsSet.UnRegisterCounter(ResCode)
	defaultOptions.metricsSet.UnRegisterSummary(ElapsedNano)
}

func TestGetServerDurationMetrics_WithNilContext(t *testing.T) {
	MetricsPromInterceptor()
	assert.Nil(t, GetServerDurationMetrics(nil))
	// unregister metrics
	clearAllMetrics()
}

func TestGetServerDurationMetrics_WithNilURL(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerDurationMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerDurationMetrics_WithNilWriter(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Writer:  nil,
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerDurationMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerDurationMetrics_HappyCase(t *testing.T) {
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})
	ctx.Request = &http.Request{
		URL: &url.URL{},
	}

	ctx.Set(rkginctx.RKEntryNameKey, rkginctx.RKEntryDefaultName)

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerDurationMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerErrorMetrics_WithNilContext(t *testing.T) {
	assert.Nil(t, GetServerErrorMetrics(nil))
}

func TestGetServerErrorMetrics_WithNilRequest(t *testing.T) {
	ctx := &gin.Context{
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerErrorMetrics_WithNilURL(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerErrorMetrics_WithNilWriter(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Writer:  nil,
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerErrorMetrics_HappyCase(t *testing.T) {
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})
	ctx.Request = &http.Request{
		URL: &url.URL{},
	}

	ctx.Set(rkginctx.RKEntryNameKey, rkginctx.RKEntryDefaultName)

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerResCodeMetrics_WithNilContext(t *testing.T) {
	assert.Nil(t, GetServerResCodeMetrics(nil))
}

func TestGetServerResCodeMetrics_WithNilRequest(t *testing.T) {
	ctx := &gin.Context{
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerResCodeMetrics_WithNilURL(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerResCodeMetrics_WithNilWriter(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{},
		Writer:  nil,
		Keys: map[string]interface{}{
			rkginctx.RKEntryNameKey: rkginctx.RKEntryDefaultName,
		},
	}

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetServerResCodeMetrics_HappyCase(t *testing.T) {
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})
	ctx.Request = &http.Request{
		URL: &url.URL{},
	}

	ctx.Set(rkginctx.RKEntryNameKey, rkginctx.RKEntryDefaultName)

	// init prom interceptor
	MetricsPromInterceptor()

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	// unregister metrics
	clearAllMetrics()
}

func TestGetValuesFromContext_WithNilContext(t *testing.T) {
	values := getValuesFromContext(nil)
	assert.Len(t, values, len(DefaultLabelKeys))
	assert.Contains(t, values, rkginctx.Realm.String)
	assert.Contains(t, values, rkginctx.Region.String)
	assert.Contains(t, values, rkginctx.AZ.String)
	assert.Contains(t, values, rkginctx.Domain.String)
	assert.Contains(t, values, rkginctx.AppVersion.String)
	assert.Contains(t, values, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
	assert.Contains(t, values, null)
	assert.Contains(t, values, null)
}

func TestGetValuesFromContext_WithNilRequest(t *testing.T) {
	ctx := &gin.Context{}
	values := getValuesFromContext(ctx)
	assert.Len(t, values, len(DefaultLabelKeys))
	assert.Contains(t, values, rkginctx.Realm.String)
	assert.Contains(t, values, rkginctx.Region.String)
	assert.Contains(t, values, rkginctx.AZ.String)
	assert.Contains(t, values, rkginctx.Domain.String)
	assert.Contains(t, values, rkginctx.AppVersion.String)
	assert.Contains(t, values, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
	assert.Contains(t, values, null)
	assert.Contains(t, values, null)
}

func TestGetValuesFromContext_WithNilURL(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Method: "unit-test-method",
		},
	}
	values := getValuesFromContext(ctx)
	assert.Len(t, values, len(DefaultLabelKeys))
	assert.Contains(t, values, rkginctx.Realm.String)
	assert.Contains(t, values, rkginctx.Region.String)
	assert.Contains(t, values, rkginctx.AZ.String)
	assert.Contains(t, values, rkginctx.Domain.String)
	assert.Contains(t, values, rkginctx.AppVersion.String)
	assert.Contains(t, values, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
	assert.Contains(t, values, "unit-test-method")
	assert.Contains(t, values, null)
	assert.Contains(t, values, null)
}

func TestGetValuesFromContext_WithNilWriter(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Method: "unit-test-method",
			URL: &url.URL{
				Path: "unit-test-path",
			},
		},
	}
	values := getValuesFromContext(ctx)
	assert.Len(t, values, len(DefaultLabelKeys))
	assert.Contains(t, values, rkginctx.Realm.String)
	assert.Contains(t, values, rkginctx.Region.String)
	assert.Contains(t, values, rkginctx.AZ.String)
	assert.Contains(t, values, rkginctx.Domain.String)
	assert.Contains(t, values, rkginctx.AppVersion.String)
	assert.Contains(t, values, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
	assert.Contains(t, values, "unit-test-method")
	assert.Contains(t, values, "unit-test-path")
	assert.Contains(t, values, null)
}

func TestGetValuesFromContext_HappyCase(t *testing.T) {
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})
	ctx.Request = &http.Request{
		Method: "unit-test-method",
		URL: &url.URL{
			Path: "unit-test-path",
		},
	}

	values := getValuesFromContext(ctx)
	assert.Len(t, values, len(DefaultLabelKeys))
	assert.Contains(t, values, rkginctx.Realm.String)
	assert.Contains(t, values, rkginctx.Region.String)
	assert.Contains(t, values, rkginctx.AZ.String)
	assert.Contains(t, values, rkginctx.Domain.String)
	assert.Contains(t, values, rkginctx.AppVersion.String)
	assert.Contains(t, values, rkentry.GlobalAppCtx.GetAppInfoEntry().AppName)
	assert.Contains(t, values, "unit-test-method")
	assert.Contains(t, values, "unit-test-path")
	assert.Contains(t, values, strconv.Itoa(200))
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

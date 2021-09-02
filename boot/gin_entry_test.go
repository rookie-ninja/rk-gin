// +build !race

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
gin:
  - name: greeter
    port: 1949
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
  - name: greeter2
    port: 2008
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
`

	serverCert = `
-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIUWVMP53O835+njsr23UZIX2KEXGYwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxCzAJBgNVBAcTAkJK
MQswCQYDVQQKEwJSSzEQMA4GA1UECxMHUksgRGVtbzETMBEGA1UEAxMKUksgRGVt
byBDQTAeFw0yMTA0MDcxMzAzMDBaFw0yNjA0MDYxMzAzMDBaMEIxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMQswCQYDVQQHEwJCSjEUMBIGA1UEAxMLZXhh
bXBsZS5uZXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARf8p/nxvY1HHUkJXZk
fFQgDtQ2CK9DOAe6y3lE21HTJ/Vi4vHNqWko9koyYgKqgUXyiq5lGAswo68KvmD7
c2L4o4GYMIGVMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTv6dUlEI6NcQBzihnzKZrxKpbnTTAfBgNV
HSMEGDAWgBRgwpYKhgfeO3p2XuX0he35caeUgTAgBgNVHREEGTAXgglsb2NhbGhv
c3SHBH8AAAGHBAAAAAAwDQYJKoZIhvcNAQELBQADggEBAByqLc3QkaGNr+QqjFw7
znk9j0X4Ucm/1N6iGIp8fUi9t+mS1La6CB1ej+FoWkSYskzqBpdIkqzqZan1chyF
njhtMsWgZYW6srXNRgByA9XS2s28+xg9owcpceXa3wG4wbnTj1emcunzSrKVFjS1
IJUjl5HWCKibnVjgt4g0s9tc8KYpXkGYl23U4FUta/07YFmtW5SDF38NWrNOe5qV
EALMz1Ry0PMgY0SDtKhddDNnNS32fz40IP0wB7a31T24eZetZK/INaIi+5SM0iLx
kfqN71xKxAIIYmuI9YwWCFaZ2+qbLIiDTbR6gyuLIQ2AfwBLZ06g939ZfSqZuP8P
oxU=
-----END CERTIFICATE-----
`

	serverKey = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPSAlU9WxdGhhmdJqg3OLmUPZlnKhejtZ2LbFNBkCTJfoAoGCCqGSM49
AwEHoUQDQgAEX/Kf58b2NRx1JCV2ZHxUIA7UNgivQzgHust5RNtR0yf1YuLxzalp
KPZKMmICqoFF8oquZRgLMKOvCr5g+3Ni+A==
-----END EC PRIVATE KEY-----
`
)

func TestWithZapLoggerEntryGin_HappyCase(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterGinEntry()

	option := WithZapLoggerEntryGin(loggerEntry)
	option(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()

	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	option := WithEventLoggerEntryGin(eventLoggerEntry)
	option(entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithInterceptorsGin_WithNilInterceptorList(t *testing.T) {
	entry := RegisterGinEntry()

	option := WithInterceptorsGin(nil)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
}

func TestWithInterceptorsGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()

	loggingInterceptor := rkginlog.Interceptor()
	metricsInterceptor := rkginmetrics.Interceptor()

	interceptors := []gin.HandlerFunc{
		loggingInterceptor,
		metricsInterceptor,
	}

	option := WithInterceptorsGin(interceptors...)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
	// should contains logging, metrics and panic interceptor
	// where panic interceptor is inject by default
	assert.Len(t, entry.Interceptors, 3)
}

func TestWithCommonServiceEntryGin_WithEntry(t *testing.T) {
	entry := RegisterGinEntry()

	option := WithCommonServiceEntryGin(NewCommonServiceEntry())
	option(entry)

	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestWithCommonServiceEntryGin_WithoutEntry(t *testing.T) {
	entry := RegisterGinEntry()

	assert.Nil(t, entry.CommonServiceEntry)
}

func TestWithTVEntryGin_WithEntry(t *testing.T) {
	entry := RegisterGinEntry()

	option := WithTVEntryGin(NewTvEntry())
	option(entry)

	assert.NotNil(t, entry.TvEntry)
}

func TestWithTVEntry_WithoutEntry(t *testing.T) {
	entry := RegisterGinEntry()

	assert.Nil(t, entry.TvEntry)
}

func TestWithCertEntryGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()
	certEntry := &rkentry.CertEntry{}

	option := WithCertEntryGin(certEntry)
	option(entry)

	assert.Equal(t, entry.CertEntry, certEntry)
}

func TestWithSWEntryGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()
	sw := NewSwEntry()

	option := WithSwEntryGin(sw)
	option(entry)

	assert.Equal(t, entry.SwEntry, sw)
}

func TestWithPortGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()
	port := uint64(1111)

	option := WithPortGin(port)
	option(entry)

	assert.Equal(t, entry.Port, port)
}

func TestWithNameGin_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()
	name := "unit-test-entry"

	option := WithNameGin(name)
	option(entry)

	assert.Equal(t, entry.EntryName, name)
}

func TestRegisterGinEntriesWithConfig_WithInvalidConfigFilePath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterGinEntriesWithConfig("/invalid-path")
}

func TestRegisterGinEntriesWithConfig_WithNilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGinEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
}

func TestRegisterGinEntriesWithConfig_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGinEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*GinEntry)
	assert.NotNil(t, greeter)
	assert.Equal(t, uint64(1949), greeter.Port)
	assert.NotNil(t, greeter.SwEntry)
	assert.NotNil(t, greeter.CommonServiceEntry)
	assert.NotNil(t, greeter.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.Len(t, greeter.Interceptors, 4)

	greeter2 := entries["greeter2"].(*GinEntry)
	assert.NotNil(t, greeter2)
	assert.Equal(t, uint64(2008), greeter2.Port)
	assert.NotNil(t, greeter2.SwEntry)
	assert.NotNil(t, greeter2.CommonServiceEntry)
	assert.NotNil(t, greeter2.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.Len(t, greeter2.Interceptors, 4)
}

func TestRegisterGinEntry_WithZapLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterGinEntry(WithZapLoggerEntryGin(loggerEntry))
	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestRegisterGinEntry_WithEventLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopEventLoggerEntry()

	entry := RegisterGinEntry(WithEventLoggerEntryGin(loggerEntry))
	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestNewGinEntry_WithInterceptors(t *testing.T) {
	loggingInterceptor := rkginlog.Interceptor()
	entry := RegisterGinEntry(WithInterceptorsGin(loggingInterceptor))
	assert.Len(t, entry.Interceptors, 2)
}

func TestNewGinEntry_WithCommonServiceEntry(t *testing.T) {
	entry := RegisterGinEntry(WithCommonServiceEntryGin(NewCommonServiceEntry()))
	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestNewGinEntry_WithTVEntry(t *testing.T) {
	entry := RegisterGinEntry(WithTVEntryGin(NewTvEntry()))
	assert.NotNil(t, entry.TvEntry)
}

func TestNewGinEntry_WithCertStore(t *testing.T) {
	certEntry := &rkentry.CertEntry{}

	entry := RegisterGinEntry(WithCertEntryGin(certEntry))
	assert.Equal(t, certEntry, entry.CertEntry)
}

func TestNewGinEntry_WithSWEntry(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterGinEntry(WithSwEntryGin(sw))
	assert.Equal(t, sw, entry.SwEntry)
}

func TestNewGinEntry_WithPort(t *testing.T) {
	entry := RegisterGinEntry(WithPortGin(1949))
	assert.Equal(t, uint64(1949), entry.Port)
}

func TestNewGinEntry_WithName(t *testing.T) {
	entry := RegisterGinEntry(WithNameGin("unit-test-greeter"))
	assert.Equal(t, "unit-test-greeter", entry.GetName())
}

func TestNewGinEntry_WithDefaultValue(t *testing.T) {
	entry := RegisterGinEntry()
	assert.True(t, strings.HasPrefix(entry.GetName(), "GinServer-"))
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Len(t, entry.Interceptors, 1)
	assert.NotNil(t, entry.Router)
	assert.Nil(t, entry.SwEntry)
	assert.Nil(t, entry.CertEntry)
	assert.False(t, entry.IsSwEnabled())
	assert.False(t, entry.IsTlsEnabled())
	assert.Nil(t, entry.CommonServiceEntry)
	assert.Nil(t, entry.TvEntry)
	assert.Equal(t, "GinEntry", entry.GetType())
}

func TestGinEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterGinEntry(WithNameGin("unit-test-entry"))
	assert.Equal(t, "unit-test-entry", entry.GetName())
}

func TestGinEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, "GinEntry", RegisterGinEntry().GetType())
}

func TestGinEntry_String_HappyCase(t *testing.T) {
	assert.NotEmpty(t, RegisterGinEntry().String())
}

func TestGinEntry_IsSWEnabled_ExpectTrue(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterGinEntry(WithSwEntryGin(sw))
	assert.True(t, entry.IsSwEnabled())
}

func TestGinEntry_IsSWEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterGinEntry()
	assert.False(t, entry.IsSwEnabled())
}

func TestGinEntry_IsTLSEnabled_ExpectTrue(t *testing.T) {
	certEntry := &rkentry.CertEntry{
		Store: &rkentry.CertStore{},
	}

	entry := RegisterGinEntry(WithCertEntryGin(certEntry))
	assert.True(t, entry.IsTlsEnabled())
}

func TestGinEntry_IsTLSEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterGinEntry()
	assert.False(t, entry.IsTlsEnabled())
}

func TestGinEntry_GetServer_HappyCase(t *testing.T) {
	entry := RegisterGinEntry()
	assert.NotNil(t, entry.Server)
	assert.NotNil(t, entry.Server.Handler)
	assert.Equal(t, "0.0.0.0:80", entry.Server.Addr)
}

func TestGinEntry_RegisterInterceptor_WithNilInterceptor(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := RegisterGinEntry()
	entry.RegisterInterceptor(nil)
}

func TestGinEntry_RegisterInterceptor_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := RegisterGinEntry()
	loggingInterceptor := rkginlog.Interceptor()
	entry.RegisterInterceptor(loggingInterceptor)
	assert.Len(t, entry.Interceptors, 2)
}

func TestGinEntry_Bootstrap_WithSwagger(t *testing.T) {
	sw := NewSwEntry(
		WithPathSw("sw"),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()))
	entry := RegisterGinEntry(
		WithNameGin("unit-test-entry"),
		WithPortGin(8080),
		WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()),
		WithSwEntryGin(sw))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.Len(t, entry.Router.Routes(), 2)

	entry.Interrupt(context.Background())
}

func TestGinEntry_Bootstrap_WithoutSwagger(t *testing.T) {
	entry := RegisterGinEntry(
		WithNameGin("unit-test-entry"),
		WithPortGin(8080),
		WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.Empty(t, entry.Router.Routes())

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestGinEntry_Bootstrap_WithoutTLS(t *testing.T) {
	entry := RegisterGinEntry(
		WithNameGin("unit-test-entry"),
		WithPortGin(8080),
		WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
}

func TestGinEntry_Shutdown_WithBootstrap(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := RegisterGinEntry(
		WithNameGin("unit-test-entry"),
		WithPortGin(8080),
		WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
}

func TestGinEntry_Shutdown_WithoutBootstrap(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := RegisterGinEntry(
		WithNameGin("unit-test-entry"),
		WithPortGin(8080),
		WithZapLoggerEntryGin(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryGin(rkentry.NoopEventLoggerEntry()))

	entry.Interrupt(context.Background())
}

func validateServerIsUp(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	if conn != nil {
		assert.Nil(t, conn.Close())
	}
}

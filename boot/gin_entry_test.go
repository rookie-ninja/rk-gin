// +build !race

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgin

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	"github.com/rookie-ninja/rk-gin/interceptor/meta"
	rkginmetrics "github.com/rookie-ninja/rk-gin/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
gin:
 - name: greeter
   port: 1949
   enabled: true
   sw:
     enabled: true
     path: "sw"
   commonService:
     enabled: true
   tv:
     enabled: true
   prom:
     enabled: true
     pusher:
       enabled: false
   interceptors:
     loggingZap:
       enabled: true
     metricsProm:
       enabled: true
     auth:
       enabled: true
       basic:
         - "user:pass"
     meta:
       enabled: true
     tracingTelemetry:
       enabled: true
     ratelimit:
       enabled: true
     timeout:
       enabled: true
     cors:
       enabled: true
     jwt:
       enabled: true
     secure:
       enabled: true
     csrf:
       enabled: true
     gzip:
       enabled: true
 - name: greeter2
   port: 2008
   enabled: true
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
 - name: greeter3
   port: 2022
   enabled: false
`
)

func TestGetGinEntry(t *testing.T) {
	// expect nil
	assert.Nil(t, GetGinEntry("entry-name"))

	// happy case
	ginEntry := RegisterGinEntry(WithNameGin("ut-gin"))
	assert.Equal(t, ginEntry, GetGinEntry("ut-gin"))

	rkentry.GlobalAppCtx.RemoveEntry("ut-gin")
}

func TestRegisterGinEntry(t *testing.T) {
	// without options
	entry := RegisterGinEntry()
	assert.NotNil(t, RegisterGinEntry())
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())

	// with options
	entry = RegisterGinEntry(
		WithZapLoggerEntryGin(nil),
		WithEventLoggerEntryGin(nil),
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithTvEntryGin(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntryGin(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntryGin(rkentry.RegisterCertEntry()),
		WithSwEntryGin(rkentry.RegisterSwEntry()),
		WithPortGin(8080),
		WithNameGin("ut-entry"),
		WithDescriptionGin("ut-desc"),
		WithPromEntryGin(rkentry.RegisterPromEntry()))

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.True(t, entry.IsSwEnabled())
	assert.True(t, entry.IsStaticFileHandlerEnabled())
	assert.True(t, entry.IsPromEnabled())
	assert.True(t, entry.IsCommonServiceEnabled())
	assert.True(t, entry.IsTvEnabled())
	assert.True(t, entry.IsTlsEnabled())

	bytes, err := entry.MarshalJSON()
	assert.NotEmpty(t, bytes)
	assert.Nil(t, err)
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestGinEntry_AddInterceptor(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterGinEntry()
	inter := rkginmeta.Interceptor()
	entry.AddInterceptor(inter)
}

func TestGinEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without enable sw, static, prom, common, tv, tls
	entry := RegisterGinEntry(WithPortGin(8080))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.Empty(t, entry.Router.Routes())

	entry.Interrupt(context.TODO())

	// with enable sw, static, prom, common, tv, tls
	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	entry = RegisterGinEntry(
		WithPortGin(8080),
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithTvEntryGin(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntryGin(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntryGin(certEntry),
		WithSwEntryGin(rkentry.RegisterSwEntry()),
		WithPromEntryGin(rkentry.RegisterPromEntry()))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.Router.Routes())

	entry.Interrupt(context.TODO())
}

func TestGinEntry_startServer_InvalidTls(t *testing.T) {
	defer assertPanic(t)

	// with invalid tls
	entry := RegisterGinEntry(
		WithPortGin(8080),
		WithCertEntryGin(rkentry.RegisterCertEntry()))
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestGinEntry_startServer_TlsServerFail(t *testing.T) {
	defer assertPanic(t)

	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	// let's give an invalid port
	entry := RegisterGinEntry(
		WithPortGin(808080),
		WithCertEntryGin(certEntry))

	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestGinEntry_startServer_ServerFail(t *testing.T) {
	defer assertPanic(t)

	// let's give an invalid port
	entry := RegisterGinEntry(
		WithPortGin(808080))

	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestRegisterGinEntriesWithConfig(t *testing.T) {
	assertNotPanic(t)

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterGinEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*GinEntry)
	assert.NotNil(t, greeter)

	greeter2 := entries["greeter2"].(*GinEntry)
	assert.NotNil(t, greeter2)

	greeter3 := entries["greeter3"]
	assert.Nil(t, greeter3)
}

func TestGinEntry_constructSwUrl(t *testing.T) {
	// happy case
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Request = &http.Request{
		Host: "8.8.8.8:1111",
	}

	path := "ut-sw"
	port := 1111

	sw := rkentry.RegisterSwEntry(rkentry.WithPathSw(path), rkentry.WithPortSw(uint64(port)))
	entry := RegisterGinEntry(WithSwEntryGin(sw), WithPortGin(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// with tls
	ctx.Request.TLS = &tls.ConnectionState{}
	assert.Equal(t, fmt.Sprintf("https://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// without swagger
	entry = RegisterGinEntry(WithPortGin(uint64(port)))
	assert.Equal(t, "N/A", entry.constructSwUrl(ctx))
}

func TestGinEntry_API(t *testing.T) {
	defer assertNotPanic(t)

	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	entry := RegisterGinEntry(
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithNameGin("unit-test-gin"))

	entry.Router.GET("ut-test")

	entry.ListApis(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestGinEntry_Req_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	entry := RegisterGinEntry(
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithNameGin("ut-gin"))

	entry.AddInterceptor(rkginmetrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut-gin", "Gin"),
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	entry.Bootstrap(context.TODO())

	entry.Req(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestGinEntry_Req_WithEmpty(t *testing.T) {
	defer assertNotPanic(t)

	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	entry := RegisterGinEntry(
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithNameGin("ut-gin"))

	entry.AddInterceptor(rkginmetrics.Interceptor(
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	entry.Bootstrap(context.TODO())

	entry.Req(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestGinEntry_TV(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterGinEntry(
		WithCommonServiceEntryGin(rkentry.RegisterCommonServiceEntry()),
		WithTvEntryGin(rkentry.RegisterTvEntry()),
		WithNameGin("ut-gin"))

	entry.AddInterceptor(rkginmetrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut-gin", "Gin")))

	entry.Bootstrap(context.TODO())

	// for /api
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)
	ctx.Params = append(ctx.Params, gin.Param{
		Key:   "item",
		Value: "/apis",
	})

	entry.TV(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	// for default
	writer = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(writer)
	ctx.Params = append(ctx.Params, gin.Param{
		Key:   "item",
		Value: "/other",
	})

	entry.TV(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func generateCerts() ([]byte, []byte) {
	// Create certs and return as []byte
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Fake cert."},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, _ := rsa.GenerateKey(rand.Reader, 4096)

	// Use CA Cert to sign a CSR and create a Public Cert
	csr := &key.PublicKey
	cert, _ := x509.CreateCertificate(rand.Reader, ca, ca, csr, key)

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(c), pem.EncodeToMemory(k)
}

func validateServerIsUp(t *testing.T, port uint64, isTls bool) {
	// sleep for 2 seconds waiting server startup
	time.Sleep(2 * time.Second)

	if !isTls {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
		if conn != nil {
			assert.Nil(t, conn.Close())
		}
		return
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	tlsConn, err := tls.Dial("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), tlsConf)
	assert.Nil(t, err)
	assert.NotNil(t, tlsConn)
	if tlsConn != nil {
		assert.Nil(t, tlsConn.Close())
	}
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

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

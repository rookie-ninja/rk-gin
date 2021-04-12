// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

//
//import (
//	"context"
//	"github.com/rookie-ninja/rk-common/common"
//	rkentry "github.com/rookie-ninja/rk-entry/entry"
//	"github.com/rookie-ninja/rk-query"
//	"github.com/stretchr/testify/assert"
//	"path"
//	"strings"
//	"testing"
//)
//
//func TestWithLogger_HappyCase(t *testing.T) {
//	logger := rkentry.NoopZapLoggerEntry()
//	entry := NewTLSEntry()
//	option := WithZapLoggerTLS(logger)
//	option(entry)
//
//	assert.Equal(t, logger, entry.zapLogger)
//}
//
//func TestWithCertFilePath_HappyCase(t *testing.T) {
//	cerFilePath := "unit-test-path"
//	entry := NewTLSEntry(
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithEventLoggerEntryTLS(rkentry.NoopEventLoggerEntry()))
//	option := WithCertFilePathTLS(cerFilePath)
//	option(entry)
//
//	assert.Equal(t, cerFilePath, entry.certFilePath)
//}
//
//func TestWithKeyFilePath_HappyCase(t *testing.T) {
//	keyFilePath := "unit-test-path"
//	entry := NewTLSEntry(
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithEventLoggerEntryTLS(rkentry.NoopEventLoggerEntry()))
//	option := WithKeyFilePathTLS(keyFilePath)
//	option(entry)
//
//	assert.Equal(t, keyFilePath, entry.keyFilePath)
//}
//
//func TestWithGenerateCert_HappyCase(t *testing.T) {
//	entry := NewTLSEntry(
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithEventLoggerEntryTLS(rkentry.NoopEventLoggerEntry()))
//	option := WithGenerateCertTLS(true)
//	option(entry)
//
//	assert.True(t, entry.generateCert)
//}
//
//func TestWithGeneratePath_HappyCase(t *testing.T) {
//	generatePath := "unit-test-path"
//	entry := NewTLSEntry()
//	option := WithGeneratePathTLS(generatePath)
//	option(entry)
//
//	assert.True(t, strings.HasSuffix(entry.generatePath, generatePath))
//}
//
//func TestNewTLSEntry_WithoutOption(t *testing.T) {
//	entry := NewTLSEntry(
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithEventLoggerEntryTLS(rkentry.NoopEventLoggerEntry()))
//	assert.NotNil(t, entry.zapLogger)
//	assert.False(t, entry.generateCert)
//	assert.Empty(t, entry.generatePath)
//	assert.Empty(t, entry.certFilePath)
//	assert.Empty(t, entry.keyFilePath)
//	assert.Equal(t, TLSEntryNameDefault, entry.entryName)
//	assert.Equal(t, TLSEntryType, entry.entryType)
//}
//
//func TestNewTLSEntry_HappyCase(t *testing.T) {
//	logger := rkentry.NoopZapLoggerEntry()
//	generatePath := "unit-test-gen-path"
//	cerFilePath := "unit-test-cert-file-path"
//	keyFilePath := "unit-test-key-file-path"
//
//	entry := NewTLSEntry(
//		WithZapLoggerTLS(logger),
//		WithCertFilePathTLS(cerFilePath),
//		WithKeyFilePathTLS(keyFilePath),
//		WithGenerateCertTLS(true),
//		WithGeneratePathTLS(generatePath))
//
//	assert.Equal(t, logger, entry.zapLogger)
//	assert.True(t, entry.generateCert)
//	assert.True(t, strings.HasSuffix(entry.generatePath, generatePath))
//	assert.Equal(t, cerFilePath, entry.certFilePath)
//	assert.Equal(t, keyFilePath, entry.keyFilePath)
//	assert.Equal(t, TLSEntryNameDefault, entry.entryName)
//	assert.Equal(t, TLSEntryType, entry.entryType)
//}
//
//func TestTLSEntry_Bootstrap_HappyCase(t *testing.T) {
//	targetPath := t.TempDir()
//	// create cert files in unit test temp dir
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerTLS(rkentry.NoopZapLoggerEntry()),
//		WithGenerateCertTLS(true),
//		WithGeneratePathTLS(targetPath))
//
//	entry.Bootstrap(context.Background())
//
//	// root CA related
//	assert.True(t, rkcommon.FileExists(path.Join(targetPath, rootCAPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(targetPath, rootCAKeyFileName)))
//	// server related
//	assert.True(t, rkcommon.FileExists(path.Join(targetPath, serverCSRFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(targetPath, serverPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(targetPath, serverKeyPEMFileName)))
//	// generated root and server config files should be removed
//	assert.False(t, rkcommon.FileExists(path.Join(targetPath, rootCAConfigFileName)))
//	assert.False(t, rkcommon.FileExists(path.Join(targetPath, rootCSCSRConfigFileName)))
//	assert.False(t, rkcommon.FileExists(path.Join(targetPath, serverConfigFileName)))
//}
//
//func TestTLSEntry_Interrupt_HappyCase(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	entry.Interrupt(context.Background())
//}
//
//func TestTLSEntry_GetName_HappyCase(t *testing.T) {
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	assert.Equal(t, TLSEntryNameDefault, entry.GetName())
//}
//
//func TestTLSEntry_GetType_HappyCase(t *testing.T) {
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	assert.Equal(t, TLSEntryType, entry.GetType())
//}
//
//func TestTLSEntry_String_HappyCase(t *testing.T) {
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	str := entry.String()
//	assert.True(t, strings.Contains(str, "name"))
//	assert.True(t, strings.Contains(str, "type"))
//	assert.True(t, strings.Contains(str, "cert_file_path"))
//	assert.True(t, strings.Contains(str, "key_file_path"))
//	assert.True(t, strings.Contains(str, "generate_cert"))
//}
//
//func TestTLSEntry_GetCertFilePath_HappyCase(t *testing.T) {
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGenerateCertTLS(true),
//		WithGeneratePathTLS(generatePath))
//	entry.Bootstrap(context.Background())
//	assert.Equal(t, path.Join(generatePath, serverPEMFileName), entry.GetCertFilePath())
//}
//
//func TestTLSEntry_GetKeyFilePath_HappyCase(t *testing.T) {
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGenerateCertTLS(true),
//		WithGeneratePathTLS(generatePath))
//	entry.Bootstrap(context.Background())
//	assert.Equal(t, path.Join(generatePath, serverKeyPEMFileName), entry.GetKeyFilePath())
//}
//
//func TestClearTLSConfigFile_WithNilEvent(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	entry.clearTLSConfigFile(nil)
//}
//
//func TestClearTLSConfigFile_WithEmptyConfigFiles(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()))
//	entry.clearTLSConfigFile(rkquery.NewEventFactory().CreateEventNoop())
//}
//
//func TestClearTLSConfigFile_HappyCase(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	event := rkquery.NewEventFactory().CreateEventNoop()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGenerateCertTLS(true),
//		WithGeneratePathTLS(t.TempDir()))
//	entry.generateCertDir(event)
//	entry.generateRootCA(event)
//	entry.generateServerCA(event)
//	entry.clearTLSConfigFile(event)
//}
//
//func TestGenerateRootCA_WithNilEvent(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGeneratePathTLS(generatePath))
//	entry.generateRootCA(nil)
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCSCSRConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAKeyFileName)))
//}
//
//func TestGenerateRootCA_HappyCase(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGeneratePathTLS(generatePath))
//	entry.generateRootCA(rkquery.NewEventFactory().CreateEventNoop())
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCSCSRConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, rootCAKeyFileName)))
//}
//
//func TestGenerateServerCA_WithNilEvent(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGeneratePathTLS(generatePath))
//	entry.generateRootCA(nil)
//	entry.generateServerCA(nil)
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverCSRFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverKeyPEMFileName)))
//}
//
//func TestGenerateServerCA_HappyCase(t *testing.T) {
//	defer func() {
//		if r := recover(); r != nil {
//			// expect panic to be called with non nil error
//			assert.True(t, false)
//		} else {
//			// this should never be called in case of a bug
//			assert.True(t, true)
//		}
//	}()
//
//	generatePath := t.TempDir()
//	entry := NewTLSEntry(
//		WithEventLoggerTLS(rkentry.NoopEventLoggerEntry()),
//		WithZapLoggerEntryTLS(rkentry.NoopZapLoggerEntry()),
//		WithGeneratePathTLS(generatePath))
//	entry.generateRootCA(rkquery.NewEventFactory().CreateEventNoop())
//	entry.generateServerCA(rkquery.NewEventFactory().CreateEventNoop())
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverConfigFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverCSRFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverPEMFileName)))
//	assert.True(t, rkcommon.FileExists(path.Join(generatePath, serverKeyPEMFileName)))
//}

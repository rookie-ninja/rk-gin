// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkgin

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/rookie-ninja/rk-common/common"
//	rkentry "github.com/rookie-ninja/rk-entry/entry"
//	"github.com/rookie-ninja/rk-query"
//	"go.uber.org/zap"
//	"io/ioutil"
//	"os"
//	"os/exec"
//	"path"
//	"strconv"
//)
//
//var (
//	originalWD, _ = os.Getwd()
//)
//
//const (
//	TLSEntryNameDefault = "rk-gin-tls"
//	TLSEntryType        = "gin-tls"
//	// root CA file name
//	rootCAConfigFileName    = "ca-config.json"
//	rootCSCSRConfigFileName = "ca-csr.json"
//	rootCAPEMFileName       = "ca.pem"
//	rootCAKeyFileName       = "ca-key.pem"
//	// server related CA file name
//	serverConfigFileName = "server.json"
//	serverCSRFileName    = "server.csr"
//	serverPEMFileName    = "server.pem"
//	serverKeyPEMFileName = "server-key.pem"
//	// default CA config content
//	rootCAConfigContent = `{
//    "signing": {
//        "default": {
//            "expiry": "43800h"
//        },
//        "profiles": {
//            "server": {
//                "expiry": "43800h",
//                "usages": [
//                    "signing",
//                    "key encipherment",
//                    "server auth"
//                ]
//            },
//            "client": {
//                "expiry": "43800h",
//                "usages": [
//                    "signing",
//                    "key encipherment",
//                    "client auth"
//                ]
//            },
//            "peer": {
//                "expiry": "43800h",
//                "usages": [
//                    "signing",
//                    "key encipherment",
//                    "server auth",
//                    "client auth"
//                ]
//            }
//        }
//    }
//}`
//	rootCACSRContent = `{
//    "CN": "RK Demo CA",
//    "key": {
//        "algo": "rsa",
//        "size": 2048
//    },
//    "names": [
//        {
//            "C": "CN",
//            "L": "BJ",
//            "O": "RK",
//            "ST": "Beijing",
//            "OU": "RK Demo"
//        }
//    ]
//}`
//	serverConfigContent = `{
//    "CN": "example.net",
//    "hosts": [
//        "localhost",
//		"127.0.0.1",
//		"0.0.0.0"
//    ],
//    "key": {
//        "algo": "ecdsa",
//        "size": 256
//    },
//    "names": [
//        {
//            "C": "CN",
//            "ST": "Beijing",
//            "L": "BJ"
//        }
//    ]
//}`
//)
//
//type BootConfigTLS struct {
//	Enabled bool `yaml:"enabled"`
//	User    struct {
//		Enabled  bool   `yaml:"enabled"`
//		CertFile string `yaml:"certFile"`
//		KeyFile  string `yaml:"keyFile"`
//	} `yaml:"user"`
//	Auto struct {
//		Enabled    bool   `yaml:"enabled"`
//		CertOutput string `yaml:"certOutput"`
//	} `yaml:"auto"`
//}
//
//type TLSEntry struct {
//	zapLogger    *rkentry.ZapLoggerEntry
//	eventLogger  *rkentry.EventLoggerEntry
//	entryName    string
//	entryType    string
//	certFilePath string
//	keyFilePath  string
//	generateCert bool
//	generatePath string
//}
//
//type TLSOption func(*TLSEntry)
//
//func WithZapLoggerEntryTLS(logger *rkentry.ZapLoggerEntry) TLSOption {
//	return func(entry *TLSEntry) {
//		if logger != nil {
//			entry.zapLogger = logger
//		}
//	}
//}
//
//func WithEventLoggerEntryTLS(logger *rkentry.EventLoggerEntry) TLSOption {
//	return func(entry *TLSEntry) {
//		if logger != nil {
//			entry.eventLogger = logger
//		}
//	}
//}
//
//func WithZapLoggerTLS(logger *rkentry.ZapLoggerEntry) TLSOption {
//	return func(entry *TLSEntry) {
//		entry.zapLogger = logger
//	}
//}
//
//func WithEventLoggerTLS(logger *rkentry.EventLoggerEntry) TLSOption {
//	return func(entry *TLSEntry) {
//		entry.eventLogger = logger
//	}
//}
//
//func WithCertFilePathTLS(subPath string) TLSOption {
//	return func(entry *TLSEntry) {
//		entry.certFilePath = subPath
//	}
//}
//
//func WithKeyFilePathTLS(subPath string) TLSOption {
//	return func(entry *TLSEntry) {
//		entry.keyFilePath = subPath
//	}
//}
//
//func WithGenerateCertTLS(generate bool) TLSOption {
//	return func(entry *TLSEntry) {
//		entry.generateCert = generate
//	}
//}
//
//func WithGeneratePathTLS(subPath string) TLSOption {
//	return func(entry *TLSEntry) {
//		if path.IsAbs(subPath) {
//			entry.generatePath = subPath
//		} else {
//			wd, err := os.Getwd()
//			if err != nil {
//				rkcommon.ShutdownWithError(err)
//			}
//			entry.generatePath = path.Join(wd, subPath)
//		}
//	}
//}
//
//func NewTLSEntry(opts ...TLSOption) *TLSEntry {
//	entry := &TLSEntry{
//		entryName:    TLSEntryNameDefault,
//		entryType:    TLSEntryType,
//		zapLogger:    rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
//		eventLogger:  rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
//		generateCert: false,
//	}
//
//	for i := range opts {
//		opts[i](entry)
//	}
//
//	return entry
//}
//
//func (entry *TLSEntry) Bootstrap(context.Context) {
//	event := entry.eventLogger.GetEventHelper().Start("bootstrap-tls-entry")
//	defer entry.eventLogger.GetEventHelper().Finish(event)
//
//	event.AddPair("tls_enabled", "true")
//	// generate tls cert files with cfssl
//	// be sure cfssl installed on local machine
//	if entry.generateCert {
//		event.AddPair("generate_cert", "true")
//		entry.generateCertDir(event)
//		entry.generateRootCA(event)
//		entry.generateServerCA(event)
//		entry.clearTLSConfigFile(event)
//
//		entry.keyFilePath = path.Join(entry.generatePath, serverKeyPEMFileName)
//		entry.certFilePath = path.Join(entry.generatePath, serverPEMFileName)
//	}
//}
//
//func (entry *TLSEntry) Interrupt(context.Context) {
//	event := entry.eventLogger.GetEventHelper().Start("interrupt-tls-entry")
//	defer entry.eventLogger.GetEventHelper().Finish(event)
//	// no op
//}
//
//func (entry *TLSEntry) GetName() string {
//	return entry.entryName
//}
//
//func (entry *TLSEntry) GetType() string {
//	return entry.entryType
//}
//
//func (entry *TLSEntry) String() string {
//
//	m := map[string]interface{}{
//		"entry_name":     entry.entryName,
//		"entry_type":     entry.entryType,
//		"cert_file_path": entry.certFilePath,
//		"key_file_path":  entry.keyFilePath,
//		"generate_cert":  strconv.FormatBool(entry.generateCert),
//		"generate_path":  entry.generatePath,
//	}
//
//	bytes, err := json.Marshal(m)
//	if err != nil {
//		entry.zapLogger.GetLogger().Warn("failed to marshal tls entry to string", zap.Error(err))
//		bytes = make([]byte, 0)
//	}
//
//	return string(bytes)
//}
//
//func (entry *TLSEntry) GetCertFilePath() string {
//	return entry.certFilePath
//}
//
//func (entry *TLSEntry) GetKeyFilePath() string {
//	return entry.keyFilePath
//}
//
//// clear ca-config.json, ca-csr.json and server.json
//func (entry *TLSEntry) clearTLSConfigFile(event rkquery.Event) {
//	if event == nil {
//		event = entry.eventLogger.GetEventFactory().CreateEventNoop()
//	}
//	CAConfigPath := path.Join(entry.generatePath, rootCAConfigFileName)
//	CACSRPath := path.Join(entry.generatePath, rootCSCSRConfigFileName)
//	ServerJSONPath := path.Join(entry.generatePath, serverConfigFileName)
//
//	if err := os.RemoveAll(CAConfigPath); err != nil {
//		entry.zapLogger.GetLogger().Warn("failed to remove ca-config.json file",
//			zap.String("tls_ca_config_path", CAConfigPath), zap.Error(err))
//		event.SetCounter("remove_tls_ca_config_error", 1)
//		event.AddErr(err)
//	}
//
//	if err := os.RemoveAll(CACSRPath); err != nil {
//		entry.zapLogger.GetLogger().Warn("failed to remove ca-csr.json file",
//			zap.String("tls_ca_csr_path", CAConfigPath), zap.Error(err))
//		event.SetCounter("remove_tls_ca_csr_error", 1)
//		event.AddErr(err)
//	}
//
//	if err := os.RemoveAll(ServerJSONPath); err != nil {
//		entry.zapLogger.GetLogger().Warn("failed to remove server.json file",
//			zap.String("tls_server_config_path", ServerJSONPath), zap.Error(err))
//		event.SetCounter("remove_server_config_error", 1)
//		event.AddErr(err)
//	}
//}
//
//// generate root CA
//func (entry *TLSEntry) generateRootCA(event rkquery.Event) {
//	if event == nil {
//		event = entry.eventLogger.GetEventFactory().CreateEventNoop()
//	}
//	rootCAConfigPath := path.Join(entry.generatePath, rootCAConfigFileName)
//	rootCACSRPath := path.Join(entry.generatePath, rootCSCSRConfigFileName)
//	rootCAPEMPath := path.Join(entry.generatePath, rootCAPEMFileName)
//	rootCAKeyPEMPath := path.Join(entry.generatePath, rootCAKeyFileName)
//
//	// create default one if ca-config.json file does not exist
//	if !rkcommon.FileExists(rootCAConfigPath) {
//		if err := ioutil.WriteFile(rootCAConfigPath, []byte(rootCAConfigContent), 0755); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to write TLS ca-config.json file",
//				zap.String("tls_root_ca_config_json", rootCAConfigPath),
//				zap.Error(err))
//			event.SetCounter("create_tls_root_ca_config_json_error", 1)
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		}
//		entry.zapLogger.GetLogger().Info("generate ca-config.json success", zap.String("path", rootCAConfigPath))
//	}
//
//	// create default one if ca-csr.json file does not exist
//	if !rkcommon.FileExists(rootCACSRPath) {
//		if err := ioutil.WriteFile(rootCACSRPath, []byte(rootCACSRContent), 0755); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to write TLS ca-csr.json file",
//				zap.String("tls_root_ca_csr_json", rootCACSRPath),
//				zap.Error(err))
//			event.SetCounter("create_tls_root_ca_csr_json_error", 1)
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		}
//		entry.zapLogger.GetLogger().Info("generate ca-csr.json success", zap.String("path", rootCACSRPath))
//	}
//
//	// create root cert request, cert and private key if files missing
//	if !rkcommon.FileExists(rootCACSRPath) ||
//		!rkcommon.FileExists(rootCAPEMPath) ||
//		!rkcommon.FileExists(rootCAKeyPEMPath) {
//		// generate root CA with bellow command
//		// cfssl gencert -initca ca-csr.json | cfssljson -bare ca -
//		// first, lets cd to cert directory
//		if err := os.Chdir(entry.generatePath); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to chdir",
//				zap.String("tls_cert_dir", entry.generatePath),
//				zap.Error(err))
//			event.AddErr(err)
//		}
//		defer func() {
//			if err := os.Chdir(originalWD); err != nil {
//				entry.zapLogger.GetLogger().Error("failed to chdir",
//					zap.String("original_wd", originalWD),
//					zap.Error(err))
//				event.AddErr(err)
//				rkcommon.ShutdownWithError(err)
//			}
//		}()
//
//		// second, run the command
//		cmd := fmt.Sprintf("cfssl gencert -initca %s| cfssljson -bare ca -", rootCSCSRConfigFileName)
//		if output, err := exec.Command("sh", "-c", cmd).Output(); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to generate cert with cfssl",
//				zap.String("cmd", cmd),
//				zap.String("output", string(output)),
//				zap.Error(err))
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		} else {
//			entry.zapLogger.GetLogger().Info("generate ca.pem, ca-key.pem success", zap.String("path", entry.generatePath))
//		}
//	}
//}
//
//// generate server CA
//func (entry *TLSEntry) generateServerCA(event rkquery.Event) {
//	serverConfigPath := path.Join(entry.generatePath, serverConfigFileName)
//	serverCSRPath := path.Join(entry.generatePath, serverCSRFileName)
//	serverPEMPath := path.Join(entry.generatePath, serverPEMFileName)
//	serverKeyPEMPath := path.Join(entry.generatePath, serverKeyPEMFileName)
//
//	// create default one if server.json file does not exist
//	if !rkcommon.FileExists(serverConfigPath) {
//		if err := ioutil.WriteFile(serverConfigPath, []byte(serverConfigContent), 0755); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to write TLS server.json file",
//				zap.String("tls_server_json_path", serverConfigPath),
//				zap.Error(err))
//			event.SetCounter("create_tls_server_json_error", 1)
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		} else {
//			entry.zapLogger.GetLogger().Info("generate server.json success", zap.String("path", serverConfigPath))
//		}
//	}
//
//	// create server cert request, cert and private key if files missing
//	if !rkcommon.FileExists(serverCSRPath) ||
//		!rkcommon.FileExists(serverPEMPath) ||
//		!rkcommon.FileExists(serverKeyPEMPath) {
//		// generate root CA with bellow command
//		// cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server server.json | cfssljson -bare server
//		// first, lets cd to cert directory
//		if err := os.Chdir(entry.generatePath); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to chdir",
//				zap.String("tls_cert_dir", entry.generatePath),
//				zap.Error(err))
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		}
//		defer func() {
//			if err := os.Chdir(originalWD); err != nil {
//				entry.zapLogger.GetLogger().Error("failed to chdir",
//					zap.String("original_wd", originalWD),
//					zap.Error(err))
//				event.AddErr(err)
//				rkcommon.ShutdownWithError(err)
//			}
//		}()
//
//		// second, run the command
//		cmd := fmt.Sprintf("cfssl gencert -ca=%s -ca-key=%s -config=%s -profile=server %s | cfssljson -bare server",
//			rootCAPEMFileName,
//			rootCAKeyFileName,
//			rootCAConfigFileName,
//			serverConfigFileName)
//		if _, err := exec.Command("sh", "-c", cmd).Output(); err != nil {
//			entry.zapLogger.GetLogger().Error("failed to generate cert with cfssl",
//				zap.String("cmd", cmd),
//				zap.Error(err))
//			event.AddErr(err)
//			rkcommon.ShutdownWithError(err)
//		} else {
//			entry.zapLogger.GetLogger().Info("generate server.csr, server.pem, server-key.pem success",
//				zap.String("path", entry.generatePath))
//		}
//	}
//}
//
//// generate certs directory under working directory
//func (entry *TLSEntry) generateCertDir(event rkquery.Event) {
//	event.AddPair("tls_cert_dir", entry.generatePath)
//	// create directory of cert
//	if err := os.MkdirAll(entry.generatePath, os.ModePerm); err != nil {
//		entry.zapLogger.GetLogger().Error("failed to mkdir",
//			zap.String("tls_cert_dir", entry.generatePath),
//			zap.Error(err))
//		event.SetCounter("create_tls_cert_dir_err", 1)
//		event.AddErr(err)
//		rkcommon.ShutdownWithError(err)
//	}
//}

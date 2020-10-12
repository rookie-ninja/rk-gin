// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_inter_logging

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	realm         = zap.String("realm", getEnvValueOrDefault("REALM", "unknown"))
	region        = zap.String("region", getEnvValueOrDefault("REGION", "unknown"))
	az            = zap.String("az", getEnvValueOrDefault("AZ", "unknown"))
	domain        = zap.String("domain", getEnvValueOrDefault("DOMAIN", "unknown"))
	appVersion    = zap.String("app_version", getEnvValueOrDefault("APP_VERSION", "latest"))
	localIP       = zap.String("local.IP", getLocalIp())
	localHostname = zap.String("local.hostname", getLocalHostname())
	appName       = "Unknown"
	eventFactory  *rk_query.EventFactory
)

const (
	RequestIdKeyLowerCase = "requestid"
	RequestIdKeyDash      = "request-id"
	RequestIdKeyUnderline = "request_id"
	RequestIdKeyDefault   = RequestIdKeyDash
	RKEventKey            = "rk-event"
)

func getEnvValueOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)

	if len(value) < 1 {
		return defaultValue
	}

	return value
}

// Get remote endpoint information set including IP, Port, NetworkType
// We will do as best as we can to determine it
// If fails, then just return default ones
func getRemoteAddressSet(ctx *gin.Context) []zap.Field {
	remoteIP := "0.0.0.0"
	remotePort := "0"

	remoteIP, remotePort, _ = net.SplitHostPort(ctx.Request.RemoteAddr)

	forwardedRemoteIP := ctx.GetHeader("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIP) > 0 {
		if forwardedRemoteIP == "::1" {
			forwardedRemoteIP = "localhost"
		}

		remoteIP = forwardedRemoteIP
	}

	if remoteIP == "::1" {
		remoteIP = "localhost"
	}

	return []zap.Field{
		zap.String("remote.IP", remoteIP),
		zap.String("remote.port", remotePort),
	}
}

// This is a tricky function
// We will iterate through all the network interfaces
// but will choose the first one since we are assuming that
// eth0 will be the default one to use in most of the case.
//
// Currently, we do not have any interfaces for selecting the network
// interface yet.
func getLocalIp() string {
	localIP := "localhost"

	// skip the error since we don't want to break RPC calls because of it
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		items := strings.Split(addr.String(), "/")
		if len(items) < 2 || items[0] == "127.0.0.1" {
			continue
		}
		if match, err := regexp.MatchString(`\d+\.\d+\.\d+\.\d+`, items[0]); err == nil && match {
			localIP = items[0]
		}
	}

	return localIP
}

func getLocalHostname() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) < 1 {
		hostname = "unknown"
	}

	return hostname
}

func GetRequestIdsFromHeader(header http.Header) []string {
	dash := header.Get(RequestIdKeyDash)
	underLine := header.Get(RequestIdKeyUnderline)
	lower := header.Get(RequestIdKeyLowerCase)

	res := make([]string, 0)

	if len(dash) > 0 {
		res = append(res, dash)
	}

	if len(underLine) > 0 {
		res = append(res, underLine)
	}

	if len(lower) > 0 {
		res = append(res, lower)
	}

	return res
}

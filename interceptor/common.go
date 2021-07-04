package rkgininter

import (
	"github.com/gin-gonic/gin"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"net"
	"strings"
)

var (
	Realm         = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	Region        = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	AZ            = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	Domain        = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	LocalIp       = zap.String("localIp", rkcommon.GetLocalIP())
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())
)

const (
	RpcEntryNameKey           = "ginEntryName"
	RpcEntryNameValue         = "gin"
	RpcEntryTypeValue         = "gin"
	RpcEventKey               = "ginEvent"
	RpcLoggerKey              = "ginLogger"
	RpcTracerKey              = "ginTracer"
	RpcSpanKey                = "ginSpan"
	RpcTracerProviderKey      = "ginTracerProvider"
	RpcPropagatorKey          = "ginPropagator"
	RpcAuthorizationHeaderKey = "authorization"
	RpcApiKeyHeaderKey        = "X-API-Key"
)

// Get remote endpoint information set including IP, Port.
// We will do as best as we can to determine it.
// If fails, then just return default ones.
func GetRemoteAddressSet(ctx *gin.Context) (remoteIp, remotePort string) {
	remoteIp, remotePort = "0.0.0.0", "0"

	if ctx == nil || ctx.Request == nil {
		return
	}

	var err error
	if remoteIp, remotePort, err = net.SplitHostPort(ctx.Request.RemoteAddr); err != nil {
		return
	}

	forwardedRemoteIp := ctx.GetHeader("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIp) > 0 {
		if forwardedRemoteIp == "::1" {
			forwardedRemoteIp = "localhost"
		}

		remoteIp = forwardedRemoteIp
	}

	if remoteIp == "::1" {
		remoteIp = "localhost"
	}

	return remoteIp, remotePort
}

func ShouldLog(ctx *gin.Context) bool {
	if ctx == nil || ctx.Request == nil {
		return false
	}

	// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
	if strings.HasPrefix(ctx.Request.URL.Path, "/rk/v1/assets") ||
		strings.HasPrefix(ctx.Request.URL.Path, "/rk/v1/tv") ||
		strings.HasPrefix(ctx.Request.URL.Path, "/sw/") {
		return false
	}

	return true
}

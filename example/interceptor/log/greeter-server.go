package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
	"github.com/rookie-ninja/rk-gin/interceptor/log/zap"
	"log"
	"net/http"
)

// In this example, we will start a new gin server with log interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []gin.HandlerFunc{
		rkginlog.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkginlog.WithEntryNameAndType("greeter", "gin"),
		//
		// Zap logger would be logged as JSON format.
		// rkginlog.WithZapLoggerEncoding(rkginlog.ENCODING_JSON),
		//
		// Event logger would be logged as JSON format.
		// rkginlog.WithEventLoggerEncoding(rkginlog.ENCODING_JSON),
		//
		// Zap logger would be logged to specified path.
		// rkginlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
		//
		// Event logger would be logged to specified path.
		// rkginlog.WithEventLoggerOutputPaths("logs/server-event.log"),
		),
	}

	// 1: Create gin server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start gin server.
func startGreeterServer(interceptors ...gin.HandlerFunc) *http.Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(interceptors...)
	router.GET("/rk/v1/greeter", Greeter)

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to listen: %v", err)
		}
	}()

	return server
}

// Response.
type GreeterResponse struct {
	Message string
}

// Handler.
func Greeter(ctx *gin.Context) {
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkginmeta.Interceptor() in server side.
	// 2: rkginctx.AddHeaderToClient(ctx, rkginctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkginctx.GetLogger(ctx).Info("Received request from client.")

	// *******************************************
	// ********** rpc-scoped event  *************
	// *******************************************
	//
	// Get rkquery.Event which would be printed as soon as request finish.
	// User can call any Add/Set/Get functions on rkquery.Event
	//
	// rkginctx.GetEvent(ctx).AddPair("rk-key", "rk-value")

	// *********************************************
	// ********** Get incoming headers *************
	// *********************************************
	//
	// Read headers sent from client.
	//
	// for k, v := range rkginctx.GetIncomingHeaders(ctx) {
	//	 fmt.Println(fmt.Sprintf("%s: %s", k, v))
	// }

	// *********************************************************
	// ********** Add headers will send to client **************
	// *********************************************************
	//
	// Send headers to client with this function
	//
	// rkginctx.AddHeaderToClient(ctx, "from-server", "value")

	// ***********************************************
	// ********** Get and log request id *************
	// ***********************************************
	//
	// RequestId will be printed on both client and server side.
	//
	// rkginctx.AddHeaderToClient(ctx, rkginctx.RequestIdKey, rkcommon.GenerateRequestId())

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}

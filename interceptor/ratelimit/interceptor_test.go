package rkginlimit

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newCtx() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	return ctx
}

func TestInterceptor(t *testing.T) {
	beforeCtx := rkmidlimit.NewBeforeCtx()
	mock := rkmidlimit.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Interceptor(rkmidlimit.WithMockOptionSet(mock))
	ctx := newCtx()
	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.New(rkerror.WithHttpCode(http.StatusUnauthorized))
	inter(ctx)
	assert.True(t, ctx.IsAborted())

	// case 2: happy case
	ctx = newCtx()
	beforeCtx.Output.ErrResp = nil
	inter(ctx)
	assert.False(t, ctx.IsAborted())
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

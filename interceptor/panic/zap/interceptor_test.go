package rkginpanic

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"os"
	"testing"
)

func TestPanicInterceptor_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	handler := PanicInterceptor()
	ctx, _ := gin.CreateTestContext(&httptest.TestResponseWriter{})

	// call interceptor
	handler(ctx)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

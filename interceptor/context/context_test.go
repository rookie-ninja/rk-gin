package rkginctx

import (
	"github.com/gin-gonic/gin"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	rkginbasic "github.com/rookie-ninja/rk-gin/interceptor/basic"
	rkginextension "github.com/rookie-ninja/rk-gin/interceptor/extension"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"os"
	"strings"
	"testing"
)

var (
	key   = "key"
	value = "value"
)

func TestAddToOutgoingHeader_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	AddToOutgoingHeader(nil, key, value)
}

func TestAddToOutgoingHeader_WithContextWithNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	AddToOutgoingHeader(&gin.Context{}, key, value)
}

func TestAddToOutgoingHeader_WithEmptyKey(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	AddToOutgoingHeader(ctx, "", value)

	assert.Equal(t, value, writer.Header().Get(""))
}

func TestAddToOutgoingHeader_WithEmptyValue(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	AddToOutgoingHeader(ctx, key, "")

	assert.Equal(t, "", writer.Header().Get(key))
}

func TestAddToOutgoingHeader_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	AddToOutgoingHeader(ctx, key, value)

	assert.Equal(t, value, writer.Header().Get(key))
}

func TestAddRequestIdToOutgoingHeader_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, SetRequestIdToOutgoingHeader(nil))
}

func TestAddRequestIdToOutgoingHeader_WithContextWithNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, SetRequestIdToOutgoingHeader(&gin.Context{}))
}

func TestAddRequestIdToOutgoingHeader_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	id := SetRequestIdToOutgoingHeader(ctx)
	assert.NotEmpty(t, id)

	assert.Equal(t, id, ctx.Writer.Header().Get(rkginextension.RequestIdHeaderKeyDefault))
}

func TestGetEvent_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	GetEvent(nil)
}

func TestGetEvent_WithoutEventInContext(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	assert.NotNil(t, GetEvent(ctx))
}

func TestGetEvent_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	event := rkquery.NewEventFactory().CreateEventNoop()
	ctx.Set(rkginbasic.RkEventKey, event)

	assert.Equal(t, event, GetEvent(ctx))
}

func TestGetLogger_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	GetLogger(nil)
}

func TestGetLogger_WithoutLoggerInContext(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	assert.NotNil(t, GetLogger(ctx))
}

func TestGetLogger_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	logger := rklogger.NoopLogger
	ctx.Set(rkginbasic.RkLoggerKey, logger)

	assert.Equal(t, logger, GetLogger(ctx))
}

func TestGetRequestIdsFromOutgoingHeader_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, GetRequestId(nil))
}

func TestGetRequestIdsFromOutgoingHeader_WithContextWithNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, GetRequestId(&gin.Context{}))
}

func TestGetRequestIdsFromOutgoingHeader_WithoutRequestId(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	assert.Empty(t, GetRequestId(ctx))
}

func TestGetRequestIdsFromOutgoingHeader_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	id := rkcommon.GenerateRequestId()

	ctx.Writer.Header().Set(rkginextension.RequestIdHeaderKeyDefault, id)

	res := GetRequestId(ctx)
	assert.Contains(t, res, id)
}

func TestGenerateRequestId_HappyCase(t *testing.T) {
	assert.NotEmpty(t, rkcommon.GenerateRequestId())
}

func TestGenerateRequestIdWithPrefix_WithEmptyPrefix(t *testing.T) {
	id := rkcommon.GenerateRequestIdWithPrefix("")
	assert.NotEmpty(t, id)
	assert.False(t, strings.HasPrefix(id, "-"))
}

func TestGenerateRequestIdWithPrefix_HappyCase(t *testing.T) {
	id := rkcommon.GenerateRequestIdWithPrefix("unit-test")
	assert.NotEmpty(t, id)
	assert.True(t, strings.HasPrefix(id, "unit-test-"))
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	code := m.Run()
	os.Exit(code)
}

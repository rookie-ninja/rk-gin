package rkginctx

import (
	"github.com/gin-gonic/gin"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	httptest "github.com/stretchr/testify/http"
	"net/http"
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

	assert.Empty(t, AddRequestIdToOutgoingHeader(nil))
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

	assert.Empty(t, AddRequestIdToOutgoingHeader(&gin.Context{}))
}

func TestAddRequestIdToOutgoingHeader_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	id := AddRequestIdToOutgoingHeader(ctx)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, ctx.Writer.Header().Get(RequestIdKeyDefault))
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
	ctx.Set(RKEventKey, event)

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
	ctx.Set(RKLoggerKey, logger)

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

	assert.Empty(t, GetRequestIdsFromOutgoingHeader(nil))
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

	assert.Empty(t, GetRequestIdsFromOutgoingHeader(&gin.Context{}))
}

func TestGetRequestIdsFromOutgoingHeader_WithoutRequestId(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	assert.Empty(t, GetRequestIdsFromOutgoingHeader(ctx))
}

func TestGetRequestIdsFromOutgoingHeader_HappyCase(t *testing.T) {
	writer := &httptest.TestResponseWriter{}
	ctx, _ := gin.CreateTestContext(writer)

	first := rkcommon.GenerateRequestId()
	second := rkcommon.GenerateRequestId()
	third := rkcommon.GenerateRequestId()

	ctx.Writer.Header().Set(RequestIdKeyDash, first)
	ctx.Writer.Header().Set(RequestIdKeyUnderline, second)
	ctx.Writer.Header().Set(RequestIdKeyLowerCase, third)

	ids := GetRequestIdsFromOutgoingHeader(ctx)
	assert.Len(t, ids, 3)

	assert.Contains(t, ids, first)
	assert.Contains(t, ids, second)
	assert.Contains(t, ids, third)
}

func TestGetRequestIdsFromIncomingHeader_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, GetRequestIdsFromIncomingHeader(nil))
}

func TestGetRequestIdsFromIncomingHeader_WithContextWithNilRequest(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	assert.Empty(t, GetRequestIdsFromIncomingHeader(&gin.Context{}))
}

func TestGetRequestIdsFromIncomingHeader_WithContextWithNilHeader(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	ctx := &gin.Context{
		Request: &http.Request{},
	}
	assert.Empty(t, GetRequestIdsFromIncomingHeader(ctx))
}

func TestGetRequestIdsFromIncomingHeader_WithoutRequestId(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{},
		},
	}

	assert.Empty(t, GetRequestIdsFromIncomingHeader(ctx))
}

func TestGetRequestIdsFromIncomingHeader_HappyCase(t *testing.T) {
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{},
		},
	}

	first := rkcommon.GenerateRequestId()
	second := rkcommon.GenerateRequestId()
	third := rkcommon.GenerateRequestId()

	ctx.Request.Header.Set(RequestIdKeyDash, first)
	ctx.Request.Header.Set(RequestIdKeyUnderline, second)
	ctx.Request.Header.Set(RequestIdKeyLowerCase, third)

	ids := GetRequestIdsFromIncomingHeader(ctx)
	assert.Len(t, ids, 3)

	assert.Contains(t, ids, first)
	assert.Contains(t, ids, second)
	assert.Contains(t, ids, third)
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

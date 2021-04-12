package rkginauth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	realm    = "unit-test-realm"
	accounts = map[string]string{
		"user": "pass",
	}
)

func TestRkGinAuth_WithNilAccounts(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	BasicAuthInterceptor(nil, realm)
}

func TestRkGinAuth_WithEmptyAccounts(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	BasicAuthInterceptor(make(map[string]string), realm)
}

func TestRkGinAuth_WithEmptyRealm(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	handler := BasicAuthInterceptor(accounts, "")
	assert.NotNil(t, handler)
}

func TestRkGinAuth_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	handler := BasicAuthInterceptor(accounts, realm)
	assert.NotNil(t, handler)
}

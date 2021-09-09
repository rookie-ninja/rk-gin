// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkginauth

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor"
	"strings"
)

// Interceptor would distinguish auth set based on.
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:     rkgininter.RpcEntryNameValue,
		EntryType:     rkgininter.RpcEntryTypeValue,
		BasicRealm:    "",
		BasicAccounts: make(map[string]bool),
		ApiKey:        make(map[string]bool),
		IgnorePrefix:  make([]string, 0),
	}

	for i := range opts {
		opts[i](set)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// optionSet which is used while initializing logging interceptor
type optionSet struct {
	EntryName     string
	EntryType     string
	BasicRealm    string
	BasicAccounts map[string]bool
	ApiKey        map[string]bool
	IgnorePrefix  []string
}

// Authorized checks permission with username and password.
func (set *optionSet) Authorized(authType, cred string) bool {
	switch authType {
	case typeBasic:
		_, ok := set.BasicAccounts[cred]
		return ok
	case typeApiKey:
		_, ok := set.ApiKey[cred]
		return ok
	}

	return false
}

// ShouldAuth determine whether auth should be checked
func (set *optionSet) ShouldAuth(ctx *gin.Context) bool {
	if ctx == nil || ctx.Request == nil || (len(set.BasicAccounts) < 1 && len(set.ApiKey) < 1) {
		return false
	}

	urlPath := ctx.Request.URL.Path

	for i := range set.IgnorePrefix {
		if strings.HasPrefix(urlPath, set.IgnorePrefix[i]) {
			return false
		}
	}

	return true
}

// Option options provided to Interceptor or optionsSet while creating
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(set *optionSet) {
		set.EntryName = entryName
		set.EntryType = entryType
	}
}

// WithBasicAuth provide basic auth credentials formed as user:pass.
// We will encode credential with base64 since incoming credential from client would be encoded.
func WithBasicAuth(realm string, cred ...string) Option {
	return func(set *optionSet) {
		for i := range cred {
			set.BasicAccounts[base64.StdEncoding.EncodeToString([]byte(cred[i]))] = true
		}

		set.BasicRealm = realm
	}
}

// WithApiKeyAuth provide API Key auth credentials.
// An API key is a token that a client provides when making API calls.
// With API key auth, you send a key-value pair to the API either in the request headers or query parameters.
// Some APIs use API keys for authorization.
//
// The API key was injected into incoming header with key of X-API-Key
func WithApiKeyAuth(key ...string) Option {
	return func(set *optionSet) {
		for i := range key {
			set.ApiKey[key[i]] = true
		}
	}
}

// WithIgnorePrefix provide paths prefix that will ignore.
// Mainly used for swagger main page and RK TV entry.
func WithIgnorePrefix(paths ...string) Option {
	return func(set *optionSet) {
		set.IgnorePrefix = append(set.IgnorePrefix, paths...)
	}
}

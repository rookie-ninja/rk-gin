package rkginauth

import (
	"encoding/base64"
	"github.com/rookie-ninja/rk-gin/interceptor"
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
		BearerToken:   make(map[string]bool),
		ApiKey:        make(map[string]bool),
	}

	for i := range opts {
		opts[i](set)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Options which is used while initializing logging interceptor
type optionSet struct {
	EntryName     string
	EntryType     string
	BasicRealm    string
	BasicAccounts map[string]bool
	BearerToken   map[string]bool
	ApiKey        map[string]bool
}

// Check permission with username and password.
func (set *optionSet) Authorized(authType, cred string) bool {
	switch authType {
	case typeBasic:
		_, ok := set.BasicAccounts[cred]
		return ok
	case typeBearer:
		_, ok := set.BearerToken[cred]
		return ok
	case typeApiKey:
		_, ok := set.ApiKey[cred]
		return ok
	}

	return false
}

type Option func(*optionSet)

// Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(set *optionSet) {
		set.EntryName = entryName
		set.EntryType = entryType
	}
}

// Provide basic auth credentials formed as user:pass.
// We will encode credential with base64 since incoming credential from client would be encoded.
func WithBasicAuth(realm string, cred ...string) Option {
	return func(set *optionSet) {
		for i := range cred {
			set.BasicAccounts[base64.StdEncoding.EncodeToString([]byte(cred[i]))] = true
		}
	}
}

// Provide bearer auth credentials.
func WithBearerAuth(token ...string) Option {
	return func(set *optionSet) {
		for i := range token {
			set.BearerToken[token[i]] = true
		}
	}
}

// Provide API Key auth credentials.
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
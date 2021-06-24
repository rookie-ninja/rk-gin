package rkginmeta

import (
	"fmt"
	"github.com/rookie-ninja/rk-gin/interceptor"
)

// Interceptor would distinguish auth set based on.
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName: rkgininter.RpcEntryNameValue,
		EntryType: rkgininter.RpcEntryTypeValue,
		Prefix:    "RK",
	}

	for i := range opts {
		opts[i](set)
	}

	if len(set.Prefix) < 1 {
		set.Prefix = "RK"
	}

	set.AppNameKey = fmt.Sprintf("X-%s-App-Name", set.Prefix)
	set.AppVersionKey = fmt.Sprintf("X-%s-App-Version", set.Prefix)
	set.AppUnixTimeKey = fmt.Sprintf("X-%s-App-Unix-Time", set.Prefix)
	set.ReceivedTimeKey = fmt.Sprintf("X-%s-Received-Time", set.Prefix)

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// options which is used while initializing extension interceptor
type optionSet struct {
	EntryName       string
	EntryType       string
	Prefix          string
	LocationKey     string
	AppNameKey      string
	AppVersionKey   string
	AppUnixTimeKey  string
	ReceivedTimeKey string
}

type Option func(*optionSet)

// Provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// Provide prefix.
func WithPrefix(prefix string) Option {
	return func(opt *optionSet) {
		opt.Prefix = prefix
	}
}

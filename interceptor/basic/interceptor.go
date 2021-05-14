package rkginbasic

import (
	"github.com/gin-gonic/gin"
	"github.com/rookie-ninja/rk-gin/interceptor/context"
)

func BasicInterceptor(opts ...Option) gin.HandlerFunc {
	set := &optionSet{
		EntryName: rkginctx.RKEntryNameValue,
		EntryType: "gin",
	}

	for i := range opts {
		opts[i](set)
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return func(ctx *gin.Context) {
		if len(ctx.GetString(rkginctx.RKEntryNameKey)) < 1 {
			ctx.Set(rkginctx.RKEntryNameKey, set.EntryName)
		}

		ctx.Next()
	}
}

func getOptionSet(ctx *gin.Context) *optionSet {
	if ctx == nil {
		return nil
	}

	entryName := ctx.GetString(rkginctx.RKEntryNameKey)
	return optionsMap[entryName]
}

var optionsMap = make(map[string]*optionSet)

// options which is used while initializing logging interceptor
type optionSet struct {
	EntryName string
	EntryType string
}

type Option func(*optionSet)

func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

package rk_gin

import (
	"github.com/golang/glog"
	"os"
	"runtime/debug"
)

func exists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func shutdownWithError(err error) {
	debug.PrintStack()
	glog.Error(err)
	os.Exit(1)
}


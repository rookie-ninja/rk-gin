// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/rookie-ninja/rk-gin/boot"
)

func main() {
	boot := rk_gin.NewGinEntries("example/boot/boot.yaml", nil, nil)
	boot["greeter"].Bootstrap()
}

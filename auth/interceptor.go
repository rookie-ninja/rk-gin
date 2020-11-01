// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_gin_inter_auth

import (
	"github.com/gin-gonic/gin"
)

// RkGinAuthZap returns a gin.HandlerFunc (middleware)
//
// Use BasicAuthForRealm from gin by default
func RkGinAuthZap(accounts gin.Accounts, realm string) gin.HandlerFunc {
	return gin.BasicAuthForRealm(accounts, realm)
}

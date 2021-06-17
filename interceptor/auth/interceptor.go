// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkginauth

import (
	"github.com/gin-gonic/gin"
)

// BasicAuthInterceptor returns a gin.HandlerFunc (middleware)
//
// Use BasicAuthForRealm from gin by default
func BasicAuthInterceptor(accounts gin.Accounts, realm string) gin.HandlerFunc {
	return gin.BasicAuthForRealm(accounts, realm)
}

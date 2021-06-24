module github.com/rookie-ninja/rk-gin-example

go 1.15

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/rookie-ninja/rk-entry v0.0.0-20210623184610-64d98c370505
	github.com/rookie-ninja/rk-gin v1.1.7-0.20210617174510-8af7876fd5ef
	github.com/rookie-ninja/rk-prom v1.0.9-0.20210623102541-1f31500c9f12
)

replace github.com/rookie-ninja/rk-gin => ../../../

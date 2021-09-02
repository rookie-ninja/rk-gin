module github.com/rookie-ninja/rk-gin-example

go 1.15

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/icza/dyno v0.0.0-20200205103839-49cb13720835 // indirect
	github.com/rookie-ninja/rk-entry v1.0.1
	github.com/rookie-ninja/rk-gin v1.2.0
	github.com/rookie-ninja/rk-prom v1.1.0
)

replace github.com/rookie-ninja/rk-gin => ../../../

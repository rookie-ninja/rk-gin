module github.com/rookie-ninja/rk-gin

go 1.14

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/markbates/pkger v0.17.1
	github.com/prometheus/client_golang v1.10.0
	github.com/rookie-ninja/rk-common v1.2.3
	github.com/rookie-ninja/rk-entry v1.0.4
	github.com/rookie-ninja/rk-logger v1.2.3
	github.com/rookie-ninja/rk-prom v1.1.4
	github.com/rookie-ninja/rk-query v1.2.4
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go v1.1.11 // indirect
	go.opentelemetry.io/contrib v1.2.0
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/jaeger v1.2.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/trace v1.2.0
	go.uber.org/ratelimit v0.2.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
)

.PHONY: all
all: gomod test lint doctoc fmt

.PHONY: lint
lint:
	@echo "[golangci-lint] Running golangci-lint..."
	@golangci-lint run 2>&1
	@echo "------------------------------------[Done]"

.PHONY: test
test:
	@echo "[test] Running go test..."
	@go test ./... -coverprofile coverage.txt 2>&1
	@go tool cover -html=coverage.txt
	@echo "------------------------------------[Done]"

.PHONY: fmt
fmt:
	@echo "[fmt] Formatting go project..."
	@gofmt -s -w . 2>&1
	@echo "------------------------------------[Done]"

.PHONY: doctoc
doctoc:
	@echo "[doctoc] Running doctoc..."
	@doctoc . 2>&1
	@echo "------------------------------------[Done]"

.PHONY: swag
swag:
	@echo "[swag] Running swag..."
	@swag init --generalInfo boot/common_service_entry.go --output boot/assets/sw/config --parseDependency 2 --propertyStrategy camelcase
	@rm -rf boot/assets/sw/config/docs.go
	@echo "------------------------------------[Done]"

.PHONY: pkger
pkger:
	@echo "[pkger] Running pkger..."
	@pkger -o boot
	@echo "------------------------------------[Done]"

.PHONY: gomod
gomod:
	@echo "[gomod] Running go mod tidy..."
	@go mod tidy
	@echo "------------------------------------[Done]"

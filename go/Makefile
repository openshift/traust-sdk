.PHONY: generate test check-drift fmt fmt-check vet lint security tools

CONTRACTS   := ../../traust-contracts
ENUMS_DIR   := $(CONTRACTS)/enums/v1
SCHEMAS_DIR := $(CONTRACTS)/schemas/v1
OUT_DIR     := ./v1/enums
ROOT        := ..
GO_CI       := python3 $(ROOT)/ci/lang/go/ci.py
export PATH := $(shell go env GOPATH)/bin:$(PATH)

generate:
	go run ./internal/generate -enums $(ENUMS_DIR) -out $(OUT_DIR) -schemas $(SCHEMAS_DIR) -types ./v1/types -validate ./v1/validate/schemas_gen.go

test:
	go test ./...

check-drift: generate
	@git diff --exit-code $(OUT_DIR) ./v1/types/ ./v1/validate/schemas_gen.go || (echo "ERROR: generated files are out of date; run 'make generate' and commit" && exit 1)

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet:
	go vet ./...

lint:
	golangci-lint run ./...

tools:
	$(GO_CI) tools

security:
	$(GO_CI) security

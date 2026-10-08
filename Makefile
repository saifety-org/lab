# CI and local checks must use pinned modules, never a developer's go.work.
export GOWORK := off
export GOFLAGS := -mod=readonly
GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT ?= $(CURDIR)/bin/golangci-lint

.PHONY: fmt-check lint lint-install ci-test ci-build

fmt-check:
	bash scripts/check-format.sh

lint-install:
	GOBIN=$(CURDIR)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: fmt-check vet
	@$(GOLANGCI_LINT) version | grep -F 'version $(GOLANGCI_LINT_VERSION:v%=%) ' >/dev/null || { echo 'Run make lint-install (requires $(GOLANGCI_LINT_VERSION))'; exit 1; }
	$(GOLANGCI_LINT) run --config .golangci.yml ./...
	$(GOLANGCI_LINT) run --config .golangci.yml --build-tags onnx ./...

.PHONY: test vet build train prepare compare regression report confusables missed-injections

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./cmd/comparison ./cmd/train ./cmd/fetch-data ./cmd/gen-confusables ./cmd/missed-injections

train:
	go run ./cmd/train -data datasets/training -out artifacts/weights.json -seed 1

prepare:
	go run ./cmd/comparison prepare

# Requires the local sAIfety DeBERTa/ONNX cache; no fallback is allowed.
compare: prepare
	go run ./cmd/train -train artifacts/comparison/train.jsonl -out artifacts/comparison/weights.json -seed 1
	go build -tags onnx -o bin/bench ./cmd/bench
	./bin/bench
	go run ./cmd/comparison report

regression:
	go test ./internal/regression -count=1 -v

report:
	go run ./cmd/comparison report

# Rebuild the runtime table into artifacts; never overwrite application sources.
confusables:
	go run ./cmd/gen-confusables

# Diagnostic corpus of known misses; it is not an independent quality benchmark.
missed-injections:
	go run ./cmd/missed-injections

ci-test:
	go test -race -count=1 -timeout=5m ./...

ci-build: build
	go build ./...
	go build -tags onnx -o bin/bench ./cmd/bench

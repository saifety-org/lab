.PHONY: test vet build train prepare compare regression report

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./cmd/comparison ./cmd/train ./cmd/fetch-data

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

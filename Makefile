default: run

PHONY: run test

run:
	@go run cmd/tfw/tfw.go -race

test:
	@go test ./... -race
.PHONY: fmt lint test fuzz_core fuzz bench race

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

test:
	go test  -v ./...

fuzz_core:
	go test -fuzz=FuzzLRU ./core

fuzz:
	go test -fuzz=FuzzPoolLRU .

bench:
	go test -bench=. -benchmem > misc/benchmark ./...

race:
	go test -race ./... -timeout 5m

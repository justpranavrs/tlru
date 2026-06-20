STRESS_PATH=./benchmark/stress.txt

.PHONY: fmt lint test fuzz_core fuzz bench race stress

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

test:
	go test  -v ./...

fuzz_core:
	go test -fuzz=FuzzLRUCore ./lrucore

fuzz:
	go test -fuzz=FuzzLRU .

bench:
	go test -bench=. -benchmem ./...

race:
	go test -race ./... -timeout 5m

stress:
	go test -bench=. -count=24 -benchmem > $(STRESS_PATH) ./...

STRESS_PATH=./benchmark/stress.txt

.PHONY: fmt lint test fuzz bench race stress

fmt:
	goimports -w .

lint:
	golangci-lint run ./...

test:
	go test  -v ./...

fuzz:
	go test -fuzz=FuzzLRUCore ./lrucore

bench:
	go test -bench=. -benchmem ./...

race:
	go test -race ./...

stress:
	go test -bench=. -count=24 -benchmem > $(STRESS_PATH) ./...


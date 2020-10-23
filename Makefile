.PHONY: build
build:
	go build -o bin/ghtl main.go

.PHONY: format
format:
	goimports -w .

.PHONY: test
test:
	go test -v ./...

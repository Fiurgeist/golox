.PHONY: install
install:
	go get ./...
	go mod tidy

.PHONY: run
run:
	go run -race ./cmd/run/main.go $(file)

.PHONY: build
build:
	go build -o golox ./cmd/run/main.go

.PHONY: help
help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  install                 get all dependencies"
	@echo "  run [file=SCRIPT]       run the interpreter"
	@echo "  build                   build executable"

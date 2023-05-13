PWD := $(shell pwd)
APP_NAME := $(shell basename ${PWD})
GOLANGCI_LINT_VERSION := 1.51.2
GOLANGCI_LINT := bin/golangci-lint-$(GOLANGCI_LINT_VERSION)/golangci-lint

all: help

## run: build and run
.PHONY: run
run: build
	@echo "🚀 Running"
	./bin/${APP_NAME}

## build: build the application
.PHONY: build
build:
	@echo "👷 Building the application"
	@go build -o bin/${APP_NAME} github.com/philiplinell/go-template/cmd/cli
	@echo "✅ Built !"

## test: run tests
.PHONY: test
test:
	@echo "🧪 Testing the application"
	@go test -race ./...

## coverage: create coverage report in HTML format
.PHONY: coverage
coverage:
	@echo "📜 Creating coverage report in HTML format"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

## lint: lint the application
.PHONY: lint
lint: ${GOLANGCI_LINT}
	@echo "🕵️  Linting code"
	@${GOLANGCI_LINT} run

## install-hooks: install git hooks
.PHONY: install-hooks
install-hooks:
	@echo "👷 Installing git hooks"
	@cp ./hooks/pre-push ./.git/hooks/pre-push

## uninstall-hooks: remove git hooks from .git/hooks
uninstall-hooks:
	@echo "👷 Uninstalling git hooks"
	@rm -f ./.git/hooks/pre-push

## expvarmon: start expvarmon tui (github.com/divan/expvarmon)
.PHONY: expvarmon
expvarmon:
	expvarmon -ports="40001"

## install-delve: Installs delve debugger
.PHONY: install-delve
install-delve:
	go install github.com/go-delve/delve/cmd/dlv@latest

## debug: Starts the delve debugger
.PHONY: debug
debug:
	dlv debug github.com/philiplinell/go-template/cmd/cli

.PHONY: help
help:
	@echo "📓 Run one of the following commands using make <command>"
	@echo
	@cat Makefile | grep "^##" | column -t -s ":" | sed "s/##/  /"
	@echo

${GOLANGCI_LINT}:
	@echo "Please install golang ci lint version ${GOLANGCI_LINT_VERSION}"
	@echo "into ${GOLANGCI_LINT}"
	@echo
	@echo "You should be able to find it here: https://github.com/golangci/golangci-lint/releases/tag/v${GOLANGCI_LINT_VERSION}"
	@exit 1


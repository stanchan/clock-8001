PROJECT_NAME := "clock-8001"
PKG := "gitlab.com/Depili/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
BINARIES := clock-bridge matrix-clock sdl-clock
GOLINT := "$(GOPATH)/bin/golint"
GIT_TAG := $(shell git describe --tags --abbrev=0)
GIT_COMMIT := $(shell git rev-list -1 HEAD)
VERSION_PKG := gitlab.com/Depili/clock-8001/clock
GO_LD_FLAGS := "-X $(VERSION_PKG).gitCommit=$(GIT_COMMIT) -X $(VERSION_PKG).gitTag=$(GIT_TAG)"

.PHONY: all dep build clean test coverage coverhtml lint

all: build

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

race: dep ## Run data race detector
	@go test -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	@go test -msan -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	@go test $(PKG_LIST) -v -coverprofile .coverage.txt
	@go tool cover -func .coverage.txt

coverhtml: ## Generate global code coverage report in HTML
	@go tool cover -html=.coverage.txt

dep: ## Get the dependencies
	@go get -v -d ./...

build: dep ## Build the binary file
	@go build -ldflags $(GO_LD_FLAGS) gitlab.com/Depili/clock-8001/cmd/clock-bridge
	@go build -ldflags $(GO_LD_FLAGS) gitlab.com/Depili/clock-8001/cmd/matrix-clock
	@go build -ldflags $(GO_LD_FLAGS) gitlab.com/Depili/clock-8001/cmd/sdl-clock

clean: ## Remove previous build
	@rm -f $(BINARIES)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

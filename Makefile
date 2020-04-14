PKG := github.com/ExchangeUnion/xud-tests

GO_BIN := ${GOPATH}/bin

GOTEST := GO111MODULE=on go test -v
GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOLIST := go list -deps $(PKG)/... | grep '$(PKG)'| grep -v '/vendor/'

COMMIT := $(shell git log --pretty=format:'%h' -n 1)
LDFLAGS := -ldflags "-X $(PKG)/build.Commit=$(COMMIT)"

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint
LINT_BIN := $(GO_BIN)/golangci-lint
LINT = $(LINT_BIN) run -v

XARGS := xargs -L 1

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

default: build

#
# Dependencies
#

$(LINT_BIN):
	@$(call print, "Fetching linter")
	go get -u $(LINT_PKG)

dependencies: $(LINT_BIN)
	go mod vendor

#
# Building
#

build:
	@$(call print, "Building xud-tests")
	$(GOBUILD) -o xud-tests $(LDFLAGS) $(PKG)

install:
	@$(call print, "Installing xud-tests")
	$(GOINSTALL) $(LDFLAGS) $(PKG)

#
# Utils
#

fmt:
	@$(call print, "Formatting source")
	gofmt -l -s -w .

lint: $(LINT_BIN)
	@$(call print, "Linting source")
	$(LINT)

.PHONY: build

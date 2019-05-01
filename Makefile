PKG := github.com/ExchangeUnion/xud-tests

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

GO_BIN := ${GOPATH}/bin
LINT_BIN := $(GO_BIN)/gometalinter.v2

HAVE_LINTER := $(shell command -v $(LINT_BIN) 2> /dev/null)

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

LINT_LIST = $(shell go list -f '{{.Dir}}' ./...)

LINT = $(LINT_BIN) \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=golint \
	--line-length=72 \
	--deadline=4m $(LINT_LIST) 2>&1 | \
	grep -v 'ALL_CAPS\|OP_' 2>&1 | \
	tee /dev/stderr

default: build

# Dependencies
$(LINT_BIN):
	@$(call print, "Fetching gometalinter.v2")
	GO111MODULE=off go get -u gopkg.in/alecthomas/gometalinter.v2

# Building
build:
	@$(call print, "Building xud-tests")
	$(GOBUILD) -o xud-tests $(PKG)

install:
	@$(call print, "Installing xud-tests")
	$(GOINSTALL) $(PKG)

# Utils
fmt:
	@$(call print, "Formatting source")
	gofmt -s -w .

lint: $(LINT_BIN)
	@$(call print, "Linting source")
	GO111MODULE=on go mod vendor
	GO111MODULE=off $(LINT_BIN) --install 1> /dev/null
	test -z "$$($(LINT))"
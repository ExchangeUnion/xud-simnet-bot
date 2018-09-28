PKG := github.com/ExchangeUnion/xud-tests

GOBUILD := go build -v
GOINSTALL := go install -v

GO_BIN := ${GOPATH}/bin
DEP_BIN := $(GO_BIN)/dep
LINT_BIN := $(GO_BIN)/gometalinter.v2

HAVE_DEP := $(shell command -v $(DEP_BIN) 2> /dev/null)
HAVE_LINTER := $(shell command -v $(LINT_BIN) 2> /dev/null)

XARGS := xargs -L 1

default: dep build

GREEN := "\\033[0;32m"
NC := "\\033[0m"

define print
	echo $(GREEN)$1$(NC)
endef

LIST := go list $(PKG)/... | grep -v '/vendor/'
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

# Dependencies
$(LINT_BIN):
	@$(call print, "Fetching gometalinter.v2")
	go get -u gopkg.in/alecthomas/gometalinter.v2

$(DEP_BIN):
	@$(call print, "Fetching dep")
	go get -u github.com/golang/dep/cmd/dep

dep: $(DEP_BIN)
	@$(call print, "Compiling dependencies")
	dep ensure -v

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
	$(LIST) | go fmt -x

lint: $(LINT_BIN)
	@$(call print, "Linting source")
	$(LINT_BIN) --install 1> /dev/null
	test -z "$$($(LINT))"

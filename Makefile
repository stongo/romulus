PROJECT = github.com/timelinelabs/romulus
REV = $$(git rev-parse --short=8 HEAD)
EXECUTABLE = "romulusd"
BINARY = cmd/romulusd/romulusd.go
IMAGE = romulusd
REMOTE_REPO = quay.io/timelinelabs/romulusd
LDFLAGS = "-s -X $(PROJECT)/romulus.SHA $(REV)"
TEST_COMMAND = godep go test

.PHONY: dep-save dep-restore test test-verbose build build-image install

all: test build build-image

help:
	@echo "Available targets:"
	@echo ""
	@echo "  dep-save"
	@echo "  dep-restore"
	@echo "  test"
	@echo "  test-verbose"
	@echo "  build"
	@echo "  build-image"
	@echo "  install"

dep-save:
	@echo "==> Saving dependencies to ./Godeps"
	@godep save ./...

dep-restore:
	@echo "==> Restoring dependencies from ./Godeps"
	@godep restore

test:
	@echo "==> Running all tests"
	@echo ""
	@$(TEST_COMMAND) ./...

test-verbose:
	@echo "==> Running all tests (verbose output)"
	@ echo ""
	@$(TEST_COMMAND) -test.v ./...

build:
	@echo "==> Building $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@godep go build -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE) $(BINARY)

build-image:
	@echo "==> Building linux binary"
	@ GOOS=linux CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE)-linux $(BINARY)
	@echo "==> Building docker image '$(IMAGE)'"
	@docker build -t $(IMAGE) .

publish: build-image
	@echo "==> publishing "

install:
	@echo "==> Installing $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@godep go install -ldflags $(LDFLAGS) $(BINARY)

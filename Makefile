.DEFAULT_GOAL := all

include .bingo/Variables.mk

TARGET     := kubelogin
OS         := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH       := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOARM      := $(if $(GOARM),$(GOARM),)
BIN         = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET)
ifeq ($(OS),windows)
  BIN = bin/$(OS)_$(ARCH)$(if $(GOARM),v$(GOARM),)/$(TARGET).exe
endif

GIT_TAG    := $(if $(GIT_TAG),$(GIT_TAG),)

LDFLAGS    := -X main.gitTag=$(GIT_TAG)

all: $(TARGET)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...

test-convert: $(TARGET)
	@echo "Running convert-kubeconfig integration tests..."
	go test ./test/integration/convert/... -tags=integration -timeout=5m

test-convert-smoke: $(TARGET)
	@echo "Running convert smoke tests..."
	go test ./test/integration/convert/... -tags=integration -run=TestConvertSmoke

test-convert-with-output: $(TARGET)
	@echo "Running convert-kubeconfig integration tests with output..."
	@mkdir -p test/integration/convert/_output
	KUBELOGIN_TEST_OUTPUT_DIR=$(PWD)/test/integration/convert/_output go test ./test/integration/convert/... -tags=integration -timeout=5m
	@echo "Converted kubeconfigs saved to: test/integration/convert/_output/"

test-convert-mixed-auth: $(TARGET)
	@echo "Running mixed authentication tests..."
	go test ./test/integration/convert/... -tags=integration -run=TestConvertMixedAuthMethods

test-convert-comprehensive: test-convert test-convert-mixed-auth
	@echo "All comprehensive convert tests completed"

test-all: test test-convert-comprehensive

$(TARGET): clean
	CGO_ENABLED=$(if $(CGO_ENABLED),$(CGO_ENABLED),0) go build -o $(BIN) -ldflags "$(LDFLAGS)"

clean:
	-rm -f $(BIN)

clean-test-output:
	-rm -rf test/integration/convert/_output

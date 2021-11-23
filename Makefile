PACKAGE  := github.com/aspenmesh/istio-vet

GOOS     = $(shell go env GOOS)
GOARCH   = $(shell go env GOARCH)
GOBIN = $(GOPATH)/bin/$(GOOS)-$(GOARCH)

GO_BUILD_FLAGS=-v

ALL_PKGS      := ./cmd/... ./pkg/...


GENERATED_GO = pkg/generated/api/v1/note.pb.go


#
# Normal build for developers. Just builds source.
#
all: go-build test

image:
	docker build -t istio-vet .


precommit: go-build image fmt


#
# Debug build, with all verification
#
debug: GO_BUILD_FLAGS=-v -race -gcflags '-N -l'
debug: go-build

#
# Remove all build artifacts
#
clean:
	rm -rf _build
	rm -rf pkg/generated

info:
	@echo "ALL_PKGS: $(ALL_PKGS)"

############################################################################
# NOTE:
#   The following targets are supporting targets for the publicly maintained
#   targets above. Publicly maintained targets above are always provided.
############################################################################

go-build: $(GENERATED_GO)
	go mod download
	GO111MODULE=on go install $(GO_BUILD_FLAGS) $(ALL_PKGS)

fmt:
	@if git diff --exit-code; then true; \
	else echo "Can't format code with local changes";false; fi
	@echo "Files that need formating:"
	@go fmt $(ALL_PKGS)
	@git diff --exit-code

test: go-test

go-test: _build/coverage.out

_build/coverage.out:
	GO111MODULE=on go install github.com/wadey/gocovmerge
	GO111MODULE=on go install github.com/onsi/ginkgo/ginkgo
	@mkdir -p $(@D)
	GO111MODULE=on ginkgo -r  \
    --randomizeAllSpecs \
    --randomizeSuites \
    --failOnPending \
    --cover \
    --trace \
    --race \
    --progress \
    --outputdir $(CURDIR)/_build/
	gocovmerge $(@D)/*.coverprofile > $@

# Generated go
pkg/generated/api/v1/note.pb.go: api/v1/note.proto
	@mkdir -p $(@D)
	protoc -I/usr/local/include -I. \
		--go_out=module=github.com/aspenmesh/istio-vet:. \
		$<

.PHONY: all test image precommit debug clean info fmt go-build go-test

# Disable builtin implicit rules
.SUFFIXES:

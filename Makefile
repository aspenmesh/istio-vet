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
all: go-build

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
	go install $(GO_BUILD_FLAGS) $(ALL_PKGS)

fmt:
	@if git diff --exit-code; then true; \
	else echo "Can't format code with local changes";false; fi
	@echo "Files that need formating:"
	@go fmt $(ALL_PKGS)
	@git diff --exit-code


# Generated go
pkg/generated/api/v1/note.pb.go: api/v1/note.proto
	@mkdir -p $(@D)
	protoc -I/usr/local/include -I. \
		-I$(GOPATH)/src \
		--go_out=:pkg/generated \
		$<

.PHONY: all

# Disable builtin implicit rules
.SUFFIXES:

SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif
.RECIPEPREFIX = >

export GO111MODULE := on
ROOTDIR := $(shell pwd)
VENDORDIR := $(ROOTDIR)/vendor
QUIET=@
VERIFYARGS ?=

GOOS ?=
GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH ?=
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY ?=
DOCKER_REGISTRY := $(if $(DOCKER_REGISTRY),"$(DOCKER_REGISTRY)/","")

# Use git tag as the image tag if present, else use latest.
IMAGE_TAG ?= $(shell git describe --always --tags 2> /dev/null || echo 'latest')

CONTROLLER_GEN_BINARY := $(GOPATH)/bin/controller-gen
CRDGEN_DIR ?= ./contrib/crd

pkgs = $(shell $(GO) list ./... | grep -v vendor)

define generate_k8s_api
	$(QUIET)bash $(VENDORDIR)/k8s.io/code-generator/generate-groups.sh \
		$(1) \
	    github.com/dgraph-io/dgraph-operator/pkg/client \
	    $(2) \
	    $(3) \
	    --go-header-file "$(ROOTDIR)/contrib/tools/codegen/custom-k8s-header-boilerplate.go.txt" \
		$(VERIFYARGS)
endef

define generate_k8s_api_all
	$(call generate_k8s_api,all,$(1),$(2))
endef

define generate_k8s_api_deepcopy
	$(call generate_k8s_api,deepcopy,$(1),$(2))
endef

generate-k8s-api:
> $(call generate_k8s_api_all,github.com/dgraph-io/dgraph-operator/pkg/apis,"dgraph.io:v1alpha1")

verify-generated-k8s-api:
> @${MAKE} -B -s VERIFYARGS=--verify-only generate-k8s-api

build:
> $(QUIET)echo "[*] Building dgraph-operator"
> $(QUIET)./contrib/scripts/build.sh

build-linux: export GOOS=linux
build-linux: export GOARCH=amd64
build-linux:
> $(QUIET)echo "[*] Building dgraph-operator"
> $(QUIET)./contrib/scripts/build.sh

format:
> $(QUIET)echo "[*] Formatting code"
> $(QUIET)$(GO) fmt $(pkgs)

govet:
> $(QUIET)echo "[*] Vetting code, checking for mistakes"
> $(QUIET)$(GO) vet $(pkgs)

check-lint:
> $(QUIET)echo "[*] Checking lint errors using golangci-lint"
> $(QUIET)golangci-lint run ./...

generate-cmdref: build
> $(QUIET)echo "[*] Generating cmdref for dgraph-operator"
> $(QUIET)./dgraph-operator cmdref --directory=./docs/cmdref

check-cmdref: build
> $(QUIET)echo "[*] Checking dgraph opeartor command line reference."
> $(QUIET)./contrib/scripts/cmdref_check.sh

fix-lint:
> $(QUIET)echo "[*] Fixing lint errors using golangci-lint"
> $(QUIET)golangci-lint run ./... --fix

docker: build-linux
> $(QUIET)echo '[*] Building docker image'
> $(QUIET)docker build --tag "${DOCKER_REGISTRY}dgraph/dgraph-operator:${IMAGE_TAG}" -f contrib/docker/operator/Dockerfile .

docker-push: docker
> $(QUIET)docker push "${DOCKER_REGISTRY}dgraph/dgraph-operator:${IMAGE_TAG}"

crdgen: $(CONTROLLER_GEN_BINARY)
> $(QUIET)echo '[*] Generating CRD definition for operator'
> $(QUIET)$(CONTROLLER_GEN_BINARY) crd paths=./pkg/apis/dgraph.io/v1alpha1 output:crd:dir=$(CRDGEN_DIR)

check-crdgen:
> $(QUIET)echo '[*] Validating generated CRD for DgraphCluster.'
> $(QUIET)./contrib/scripts/crdgen_check.sh

$(CONTROLLER_GEN_BINARY):
> $(QUIET)go install sigs.k8s.io/controller-tools/cmd/controller-gen

e2e-tests:
> $(QUIET)echo '[*] Running E2E tests for dgraph-operator'
> $(QUIET)./tests/e2e.sh

.PHONY: build format govet fix-lint check-lint generate-cmdref check-cmdref \
	generate-k8s-api verify-generated-k8s-api docker docker-push \
	crd-gen

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

GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath

# SET DOCKER_REGISTRY to change the docker registry
DOCKER_REGISTRY := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY),localhost:5000)
IMAGE_TAG ?= latest

pkgs = $(shell $(GO) list ./cmd/... | grep -v vendor)
pkgs += $(shell $(GO) list ./pkg/...)

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

format:
> $(QUIET)echo "[*] Formatting code"
> $(QUIET)$(GO) fmt $(pkgs)

govet:
> $(QUIET)echo "[*] Vetting code, checking for mistakes"
> $(QUIET)$(GO) vet $(pkgs)

check-lint:
> $(QUIET)echo "[*] Checking lint errors using golangci-lint"
> $(QUIET)golangci-lint run ./cmd/...

generate-cmdref: build
> $(QUIET)echo "[*] Generating cmdref for dgraph-operator"
> $(QUIET)./dgraph-operator cmdref --directory=./docs/cmdref

check-cmdref: build
> $(QUIET)echo "[*] Checking dgraph opeartor command line reference."
> $(QUIET)./contrib/scripts/cmdref_check.sh

fix-lint:
> $(QUIET)echo "[*] Fixing lint errors using golangci-lint"
> $(QUIET)golangci-lint run ./cmd/... --fix

docker: build
> $(QUIET)echo '[*] Building docker image'
> $(QUIET)docker build --tag "${DOCKER_REGISTRY}/dgraph-io/dgraph-operator:${IMAGE_TAG}" -f contrib/docker/operator/Dockerfile .

docker-push: docker
> $(QUIET)docker push "${DOCKER_REGISTRY}/dgraph-io/dgraph-operator:${IMAGE_TAG}"

.PHONY: build format govet fix-lint check-lint generate-cmdref check-cmdref generate-k8s-api verify-generated-k8s-api docker docker-push

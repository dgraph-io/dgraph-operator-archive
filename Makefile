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

GO := go
QUIET=@
pkgs = $(shell $(GO) list ./... | grep -v vendor)

build:
> $(QUIET)echo "[*] Building dgraph-operator"
> $(QUIET)./contrib/scripts/build.sh

format:
> $(QUIET)echo "[*] Formatting code"
> $(QUIET)$(GO) fmt $(pkgs)

govet:
> $(QUIET)echo "[*] Vetting code, checking for mistakes"
> $(QUIET)$(GO) vet $(pkgs)

check_lint:
> $(QUIET)echo "[*] Checking lint errors using golangci-lint"
> $(QUIET)golangci-lint run ./cmd/...

generate_cmdref: build
> $(QUIET)echo "[*] Generating cmdref for dgraph-operator"
> $(QUIET)./dgraph-operator cmdref --directory=./docs/cmdref

check_cmdref: build
> $(QUIET)echo "[*] Checking dgraph opeartor command line reference."
> $(QUIET)./contrib/scripts/cmdref_check.sh

.PHONY: build format govet check_lint generate_cmdref check_cmdref

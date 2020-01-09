#!/bin/bash

set -euo pipefail

CWD=${PWD}

export CGO_ENABLED=0

GO_FLAGS=${GO_FLAGS:-"-tags netgo"}
GO_CMD=${GO_CMD:-"build"}
VERBOSE=${VERBOSE:-}
BUILD_NAME="dgraph-operator"

REPO_PATH="github.com/dgraph-io/dgraph-operator"
API_VERSION="v1alpha1"
OPERATOR_VERSION=$(git describe --always --tags 2> /dev/null || echo 'unknown')
BUILD_DATE=${BUILD_DATE:-$( git log -1 --format='%ci' )}
REVISION=$(git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown')
GO_VERSION=$(go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

ldseparator="="

ldflags="
  -X '${REPO_PATH}/version.APIVersion${ldseparator}${API_VERSION}'
  -X '${REPO_PATH}/version.OperatorVersion${ldseparator}${OPERATOR_VERSION}'
  -X '${REPO_PATH}/version.CommitSHA${ldseparator}${REVISION}'
  -X '${REPO_PATH}/version.Branch${ldseparator}${BRANCH}'
  -X '${REPO_PATH}/version.CommitTimestamp${ldseparator}${BUILD_DATE}'
  -X '${REPO_PATH}/version.GoVersion${ldseparator}${GO_VERSION}'"

if [ -n "$VERBOSE" ]; then
  echo "Building with -ldflags $ldflags"
fi

GOBIN=$PWD go "${GO_CMD}" -o "${BUILD_NAME}" ${GO_FLAGS} -ldflags "${ldflags}" "${REPO_PATH}/cmd/operator"

echo "[*] Build Complete."
exit 0

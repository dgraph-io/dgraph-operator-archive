#!/bin/bash

set -euo pipefail

CWD=${PWD}

GO_FLAGS=${GO_FLAGS:-"-tags netgo"}
GO_CMD=${GO_CMD:-"build"}
BUILD_DATE=${BUILD_DATE:-$( date +%Y%m%d-%H:%M:%S )}
VERBOSE=${VERBOSE:-}
BUILD_NAME="dgraph-operator"

REPO_PATH="github.com/dgraph-io/dgraph-operator"
VERSION="v1alpha1"
REVISION=$(git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null || echo 'unknown')
GO_VERSION=$(go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

# go 1.4 requires ldflags format to be "-X key value", not "-X key=value"
# ldseparator here is for cross compatibility between go versions
ldseparator="="
if [ "${GO_VERSION:0:3}" = "1.4" ]; then
    ldseparator=" "
fi

ldflags="
  -X ${REPO_PATH}/version.Version${ldseparator}${VERSION}
  -X ${REPO_PATH}/version.Revision${ldseparator}${REVISION}
  -X ${REPO_PATH}/version.Branch${ldseparator}${BRANCH}
  -X ${REPO_PATH}/version.BuildDate${ldseparator}${BUILD_DATE}
  -X ${REPO_PATH}/version.GoVersion${ldseparator}${GO_VERSION}"

if [ -n "$VERBOSE" ]; then
  echo "Building with -ldflags $ldflags"
fi

GOBIN=$PWD go "${GO_CMD}" -o "${BUILD_NAME}" ${GO_FLAGS} -ldflags "${ldflags}" "${REPO_PATH}/cmd/operator"

echo "[*] Build Complete."
exit 0

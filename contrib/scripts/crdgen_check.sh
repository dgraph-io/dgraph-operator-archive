#!/bin/bash

set -euo pipefail

CRD_DIR=./contrib/crd/
CRDFILE_NAME=dgraph.io_dgraphclusters.yaml

TMP_DIR=`mktemp -d`

trap 'rm -rf $TMP_DIR' EXIT INT TERM

CRDGEN_DIR=$TMP_DIR make crdgen

if ! $(diff ${CRD_DIR}/${CRDFILE_NAME} ${TMP_DIR}/${CRDFILE_NAME}); then
  echo "Detected a difference in CRD definition"
  echo "diff: `diff ${CRD_DIR}/${CRDFILE_NAME} ${TMP_DIR}/${CRDFILE_NAME}`"
  echo "Please rerun 'make crdgen' and commit your changes"
  exit 1
fi

echo "[*] CRD definition is up to date."

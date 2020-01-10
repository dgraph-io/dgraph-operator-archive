#!/bin/bash

set -euo pipefail

DOCS_DIR=./docs
CMDREF_DIR=${DOCS_DIR}/cmdref

# Create a temp directory to generate the cmdref into for diff checking
TMP_DIR=`mktemp -d`

# remove temp directory if interrupted or on exit.
trap 'rm -rf $TMP_DIR' EXIT INT TERM

./dgraph-operator cmdref --directory=${TMP_DIR}

if ! $(diff -r ${CMDREF_DIR} ${TMP_DIR}); then
  echo "Detected a difference in the cmdref directory"
  echo "diff -r: `diff -r ${CMDREF_DIR} ${TMP_DIR}`"
  echo "Please rerun 'make generate-cmdref' and commit your changes"
  exit 1
fi

echo "[*] dgraph-operator cmdref is up to date."

#!/usr/bin/env bash

DOCKER_BIN=${DOCKER_BIN:-docker}
KUBECTL_BIN=${KUBECTL_BIN:-kubectl}
KIND_BIN=${KIND_BIN:-kind}
GINKGO_BIN=${GINKGO_BIN:-ginkgo}

KIND_VERSION=${KIND_VERSION:-0.7.0}

# The info command should succeed if the docker daemon is running.
function check_docker() {
    $DOCKER_BIN info > /dev/null 2>&1 || { 
        echo "[-] Docker daemon is not running, try starting it using 'systemctl start docker'"
        exit 1
    }
}

function check_kubectl() {
    if ! test -x "$(command -v $KUBECTL_BIN)"; then
        echo "[*] kubectl is either not installed or is not it path."
        exit 1
    fi
}

function check_kind() {
    if test -x "$(command -v $KIND_BIN)"; then
        [[ "$($KIND_BIN --version 2>&1 | cut -d ' ' -f 3)" == "$KIND_VERSION" ]]
        return
    fi

    echo "[-] kind is either not installed or is not in the path."
    exit 1
}

function check_ginkgo() {
    if ! test -x "$(command -v $GINKGO_BIN)"; then
        echo "[-] gingko is either not installed or is not in the path."
        exit 1
    fi
}

#!/usr/bin/env bash

set -o nounset
set -euo pipefail

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

TESTS_DIR=$ROOT/tests

function print_help() {
    echo "End To End test run script for dgraph-operator."
}

if [ $# -ge 1 ] && [ -n "$1" ]
then
    opt_key="$1"
    case $opt_key in
        -h|--help)
            print_help
            exit 0
            ;;
    esac

    # Pass the argument -- and shift to the next one
    if [ "${1:-}" == "--" ]; then
        shift
    fi
fi

source "${ROOT}/tests/helpers.sh"

check_docker
check_kind
check_kubectl
check_ginkgo

DOCKER_REGISTRY=${DOCKER_REGISTRY:-localhost:5000}
IMAGE_TAG=${IMAGE_TAG:-latest}
KUBECONFIG=${KUBECONFIG:-~/.kube/config}
KUBE_WORKERS=${KUBE_WORKERS:-1}
CLUSTER_NAME=${CLUSTER_NAME:-dgraph-operator}
SKIP_BUILD=${SKIP_BUILD:-}
REUSE_CLUSTER=${REUSE_CLUSTER:-}
SKIP_TEARDOWN_CLUSTER=${SKIP_TEARDOWN_CLUSTER:-}

echo "DOCKER_REGISTRY: $DOCKER_REGISTRY"
echo "IMAGE_TAG: $IMAGE_TAG"
echo "KUBECONFIG: $KUBECONFIG"
echo "KUBE_WORKERS: $KUBE_WORKERS"
echo "CLUSTER_NAME: $CLUSTER_NAME"
echo "SKIP_BUILD: $SKIP_BUILD"
echo "REUSE_CLUSTER: $REUSE_CLUSTER"
echo "SKIP_TEARDOWN_CLUSTER: $SKIP_TEARDOWN_CLUSTER"

# https://github.com/kubernetes-sigs/kind/releases/tag/v0.7.0
# This e2e testing script uses kind to create kubernetes cluster with kindest
# images and run the integration tests there.
declare -A kube_images
kube_images["v1.17.0"]="kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62"
kube_images["v1.16.4"]="kindest/node:v1.16.4@sha256:b91a2c2317a000f3a783489dfb755064177dbc3a0b2f4147d50f04825d016f55"

function build_image() {
    if [ -n "${SKIP_BUILD}" ]; then
        echo "[*] Skipping operator image build."
        return
    fi
    DOCKER_REGISTRY=$DOCKER_REGISTRY IMAGE_TAG=$IMAGE_TAG make docker
}

function run_kind_cluster_tests() {
    echo "[*] Setting up kind cluster"
    
    KIND_KUBE_VERSION=$1
    KIND_NODE_IMAGE=$2

    if $KIND_BIN get clusters | grep $CLUSTER_NAME &>/dev/null; then
        echo "[+] The cluster($KIND_KUBE_VERSION) is already running"

        if [ -n "$REUSE_CLUSTER" ]; then
            echo "[-] Not reusing existing cluster, exitting..."
            exit 1
        fi
    else
        tmp=$(mktemp)
        cat <<EOF > $tmp
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOF

        for ((i = 1; i <= $KUBE_WORKERS; i++)) {
            cat <<EOF >> $tmp
- role: worker
EOF
        }

        $KIND_BIN create cluster \
            --name $CLUSTER_NAME \
            --image $KIND_NODE_IMAGE \
            --config $tmp \
            --wait 120s
    fi

    echo "[*] Loading docker image into the kind cluster"
    $KIND_BIN load docker-image \
        $DOCKER_REGISTRY/dgraph/dgraph-operator:$IMAGE_TAG \
        --name $CLUSTER_NAME

    ginkgo -v -- \
        -kubeconfig $KUBECONFIG \
        -operator-image $DOCKER_REGISTRY/dgraph/dgraph-operator:$IMAGE_TAG \
        -context kind-$CLUSTER_NAME

    echo "[*] Cleaning up the cluster($KIND_NODE_IMAGE)"
    $KIND_BIN delete cluster --name $CLUSTER_NAME
}

function run_tests() {
    cd $TESTS_DIR

    echo "[*] Running operator tests on the kind cluster."
    for v in ${!kube_images[*]}; do
        echo "[*] Working on cluster version: ${v}"
        KIND_NODE_IMAGE=${kube_images["${v}"]}

        run_kind_cluster_tests $v $KIND_NODE_IMAGE
    done
}

build_image
run_tests


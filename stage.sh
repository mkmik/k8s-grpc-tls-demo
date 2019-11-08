#!/bin/sh

set -e

function usage() {
  echo "usage: $0 [args] --image-prefix <image-prefix> [build] [apply] [diff]"
  exit 1
}

cd $(dirname $0)
mkdir -p staging
cd staging

###

image_prefix=""

args=""
cnt=0
while [[ $# -gt 0 ]]; do
  ((cnt=cnt+1))
  case "$1" in
    --image-prefix) shift; image_prefix="$1" ;;
    build) do_build=1 ;;
    apply) do_apply=1 ;;
    diff)  do_diff=1 ;;
    --)
      shift
      args="${args} ${@}"
      break
    ;;
     *)
      args="${args} ${1}"
      ((cnt=cnt-1))
    ;;
  esac

  shift
done

if [ "${cnt}" == 0 -o -z "${image_prefix}" ]; then
  usage
fi

###

if [ ! -f kustomization.yaml ]; then
  kustomize create --resources ../manifests/
fi

###

if [ ! -z "${do_build}" ]; then
  for i in client server; do
    docker build .. --target=${i} -t ${image_prefix}-${i}
    docker push ${image_prefix}-${i}
    digest=$(docker inspect ${image_prefix}-${i} | jq -r '.[0].RepoDigests[0]')
    kustomize edit set image replaceme-${i}=$digest
  done
fi

function kube() {
  kustomize build . | kubectl $1 -f - ${args}
}

if [ ! -z "${do_diff}" ]; then
  (! kube diff) || false # only continue if there is any diff
fi

if [ ! -z "${do_apply}" ]; then
  kube apply
fi

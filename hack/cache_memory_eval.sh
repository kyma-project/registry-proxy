#!/usr/bin/env bash

# Measures how much memory a registry-proxy manager spends caching objects it does
# not manage. See issue #139.
#
# Every type a controller-runtime manager watches or reads through its cached client
# is backed by an informer that holds every object of that type in memory. Two
# managers were affected:
#   - operator:   read Secrets by key through the cached client -> cluster-wide Secret cache
#   - controller: Owns(&Pod{})                                  -> cluster-wide Pod cache
# Both pods are capped at 128Mi, so a cluster full of unrelated Secrets/Pods OOM-kills
# them. This eval fills the cluster with dummy objects of the chosen dimension and
# reports how much the manager grows because of them. A scoped cache stays flat.
#
# Usage: the module has to be deployed and both managers Ready.
#   DIMENSION=secret ./hack/cache_memory_eval.sh   # exercises the operator
#   DIMENSION=pod    ./hack/cache_memory_eval.sh   # exercises the controller
#   DIMENSION=pod COUNT=1500 THRESHOLD_MB=15 ./hack/cache_memory_eval.sh

set -o errexit
set -o nounset
set -o pipefail

DIMENSION="${DIMENSION:-secret}"
CONTEXT="${CONTEXT:-}"
NAMESPACE="${NAMESPACE:-kyma-system}"
EVAL_NAMESPACE="${EVAL_NAMESPACE:-cache-eval}"
# a scoped cache should stay flat, an unrestricted one grows with the payload of every object
THRESHOLD_MB="${THRESHOLD_MB:-20}"
BATCH_SIZE="${BATCH_SIZE:-200}"

case "${DIMENSION}" in
  secret)
    DEPLOYMENT="${DEPLOYMENT:-registry-proxy-operator}"
    COUNT="${COUNT:-2000}"
    SECRET_SIZE_KB="${SECRET_SIZE_KB:-16}"
    ;;
  pod)
    DEPLOYMENT="${DEPLOYMENT:-registry-proxy-controller}"
    # the controller OOM'd between 500 and 1000 pods in issue #139
    COUNT="${COUNT:-1500}"
    ;;
  *)
    echo "DIMENSION must be 'secret' or 'pod', got '${DIMENSION}'" >&2
    exit 2
    ;;
esac

kube() {
  if [ -n "${CONTEXT}" ]; then
    kubectl --context "${CONTEXT}" "$@"
  else
    kubectl "$@"
  fi
}

log() {
  echo ""
  echo "### $1"
}

# managerMemoryMi prints the working set of the target manager Pod in MiB. Pods are
# resolved from the Deployment's own selector so this does not depend on a hand-copied
# label (the operator and controller pods share the control-plane label).
managerMemoryMi() {
  local selector
  selector=$(kube -n "${NAMESPACE}" get deploy "${DEPLOYMENT}" \
    -o jsonpath='{.spec.selector.matchLabels}' 2>/dev/null |
    tr -d '{}"' | tr ',' '\n' | paste -sd, - || true)
  if [ -z "${selector}" ]; then
    echo "cannot read selector of deploy/${DEPLOYMENT} in ${NAMESPACE}" >&2
    return 1
  fi

  local output waited=0
  while [ "${waited}" -lt 120 ]; do
    output=$(kube top pod -n "${NAMESPACE}" -l "${selector}" --no-headers 2>/dev/null || true)
    if [ -n "${output}" ]; then
      echo "${output}" | awk '{gsub(/Mi/, "", $3); print $3; exit}'
      return 0
    fi
    sleep 5
    waited=$((waited + 5))
  done
  echo "metrics for deploy/${DEPLOYMENT} are not available" >&2
  return 1
}

# peakMemoryMi samples the manager memory a few times and keeps the highest value,
# because the metrics pipeline reports with a delay.
peakMemoryMi() {
  local peak=0 sample
  for _ in 1 2 3 4; do
    sample=$(managerMemoryMi)
    if [ "${sample}" -gt "${peak}" ]; then
      peak=${sample}
    fi
    sleep 20
  done
  echo "${peak}"
}

# createSecrets fills the eval namespace with dummy Secrets. They carry no
# registry-proxy label, so a correctly scoped operator cache must never hold them.
createSecrets() {
  local payload
  payload=$(head -c $((SECRET_SIZE_KB * 1024)) /dev/urandom | base64 | tr -d '\n')

  local created=0 batchEnd index manifest
  while [ "${created}" -lt "${COUNT}" ]; do
    manifest=""
    batchEnd=$((created + BATCH_SIZE))
    if [ "${batchEnd}" -gt "${COUNT}" ]; then
      batchEnd=${COUNT}
    fi
    index=${created}
    while [ "${index}" -lt "${batchEnd}" ]; do
      manifest="${manifest}
---
apiVersion: v1
kind: Secret
metadata:
  name: filler-${index}
  namespace: ${EVAL_NAMESPACE}
type: Opaque
data:
  payload: ${payload}"
      index=$((index + 1))
    done
    echo "${manifest}" | kube apply -f - >/dev/null
    created=${batchEnd}
    echo "  created ${created}/${COUNT} secrets"
  done
}

# createPods fills the eval namespace with dummy Pods that carry NO
# registry-proxy.kyma-project.io/managed-by label, so a correctly scoped controller
# cache must never hold them. A nodeSelector for a label no node carries keeps them
# Pending: they never consume node resources but still populate the Pod informer's
# list/watch on the API server, which is what the unscoped cache used to hold.
createPods() {
  local created=0 batchEnd index manifest
  while [ "${created}" -lt "${COUNT}" ]; do
    manifest=""
    batchEnd=$((created + BATCH_SIZE))
    if [ "${batchEnd}" -gt "${COUNT}" ]; then
      batchEnd=${COUNT}
    fi
    index=${created}
    while [ "${index}" -lt "${batchEnd}" ]; do
      manifest="${manifest}
---
apiVersion: v1
kind: Pod
metadata:
  name: filler-${index}
  namespace: ${EVAL_NAMESPACE}
spec:
  nodeSelector:
    cache-eval/unschedulable: \"true\"
  terminationGracePeriodSeconds: 0
  containers:
    - name: pause
      image: registry.k8s.io/pause:3.9"
      index=$((index + 1))
    done
    echo "${manifest}" | kube apply -f - >/dev/null
    created=${batchEnd}
    echo "  created ${created}/${COUNT} pods"
  done
}

createLoad() {
  case "${DIMENSION}" in
    secret) createSecrets ;;
    pod) createPods ;;
  esac
}

cleanup() {
  log "removing the dummy ${DIMENSION}s"
  kube delete namespace "${EVAL_NAMESPACE}" --ignore-not-found --wait=false >/dev/null
}
trap cleanup EXIT

log "restarting deploy/${DEPLOYMENT} to measure from a clean baseline"
kube -n "${NAMESPACE}" rollout restart "deploy/${DEPLOYMENT}" >/dev/null
kube -n "${NAMESPACE}" rollout status "deploy/${DEPLOYMENT}" --timeout=180s

log "measuring the baseline memory"
BASELINE=$(peakMemoryMi)
echo "  baseline: ${BASELINE} MiB"

if [ "${DIMENSION}" = "secret" ]; then
  log "creating ${COUNT} secrets of ${SECRET_SIZE_KB} KiB in ${EVAL_NAMESPACE}"
else
  log "creating ${COUNT} unmanaged pods in ${EVAL_NAMESPACE}"
fi
kube create namespace "${EVAL_NAMESPACE}" --dry-run=client -o yaml | kube apply -f - >/dev/null
createLoad

log "measuring the memory with the dummy ${DIMENSION}s in the cluster"
LOADED=$(peakMemoryMi)
echo "  with ${COUNT} ${DIMENSION}s: ${LOADED} MiB"

GROWTH=$((LOADED - BASELINE))

log "result"
echo "  dimension:  ${DIMENSION} (deploy/${DEPLOYMENT})"
echo "  count:      ${COUNT}"
echo "  baseline:   ${BASELINE} MiB"
echo "  loaded:     ${LOADED} MiB"
echo "  growth:     ${GROWTH} MiB (threshold ${THRESHOLD_MB} MiB)"

if [ "${GROWTH}" -gt "${THRESHOLD_MB}" ]; then
  echo ""
  echo "FAIL: the ${DEPLOYMENT} caches ${DIMENSION}s it does not manage"
  exit 1
fi

echo ""
echo "PASS: ${DEPLOYMENT} memory does not follow the number of ${DIMENSION}s in the cluster"

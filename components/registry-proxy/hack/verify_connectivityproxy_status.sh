#!/bin/bash

function get_connectivityproxy_status () {
  echo "checking connectivity proxy status"

  kubectl wait --for=jsonpath='{.status.state}'=Ready -n kyma-system connectivityproxy connectivity-proxy --timeout=90s

  sleep 5

  kubectl wait --for=jsonpath='{.status.state}'=Ready -n kyma-system connectivityproxy connectivity-proxy --timeout=60s || {
    echo "connectivityproxy resource did not reach the Ready state"
    kubectl get all --all-namespaces
    exit 1
  }

  echo "connectivityproxy reached the Ready state"
  return 0
}

get_connectivityproxy_status
#!/usr/bin/env bash

# standard bash error handling
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

MODULE_VERSION=${MODULE_VERSION?"Define MODULE_VERSION env"} # module version used to set common labels
PROJECT_ROOT=${PWD}
CONFIG=${PWD}/config

echo "ensure helm..."
PROJECT_ROOT=${PROJECT_ROOT} make -C ${PROJECT_ROOT} helm

echo "upgrade helm chart..."
cd ${CONFIG}/registry-proxy && yq --inplace ".version=\"${MODULE_VERSION}\"" Chart.yaml && yq --inplace ".appVersion=\"${MODULE_VERSION}\"" Chart.yaml
cd ${CONFIG}/registry-proxy && yq --inplace ".global.images.registry_proxy.version=\"${MODULE_VERSION}\"" values.yaml && yq --inplace ".global.images.connection.version=\"${MODULE_VERSION}\"" values.yaml
cd ${CONFIG}/operator && yq --inplace ".version=\"${MODULE_VERSION}\"" Chart.yaml && yq --inplace ".appVersion=\"${MODULE_VERSION}\"" Chart.yaml

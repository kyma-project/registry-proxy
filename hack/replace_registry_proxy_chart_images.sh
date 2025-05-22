#!/bin/bash

# if you only need replace images with version set to "main" specify "main-only" argument
REPLACE_SCOPE=$1
# IMG_DIRECTORY can be omitted
REQUIRED_ENV_VARIABLES=('IMG_VERSION' 'PROJECT_ROOT')
for VAR in "${REQUIRED_ENV_VARIABLES[@]}"; do
  if [[ -z "${!VAR}" ]]; then
    echo "${VAR} is undefined"
    exit 1
  fi
done

MAIN_ONLY_SELECTOR=""
if [[ ${REPLACE_SCOPE} == "main-only" ]]; then
  MAIN_ONLY_SELECTOR="| select(.version == \"main\")"
fi

VALUES_FILE=${PROJECT_ROOT}/config/registry-proxy/values.yaml

if [[ ${PURPOSE} == "local" ]]; then
  echo "Changing container registry to k3d-kyma-registry.localhost:5000"
  # TOOD: why won't it work with k3d registry path?
  yq -i '.global.containerRegistry.path=""' "${VALUES_FILE}"
fi

IMAGES_SELECTOR=".global.images[] ${MAIN_ONLY_SELECTOR}"
yq -i "(${IMAGES_SELECTOR} | .directory) = \"${IMG_DIRECTORY}\"" "${VALUES_FILE}"
yq -i "(${IMAGES_SELECTOR} | .version) = \"${IMG_VERSION}\"" "${VALUES_FILE}"
echo "==== Local Changes ===="
yq '.global.images' "${VALUES_FILE}"
yq '.global.containerRegistry' "${VALUES_FILE}"
echo "==== End of Local Changes ===="

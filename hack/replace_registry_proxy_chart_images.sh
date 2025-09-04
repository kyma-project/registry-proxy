#!/bin/bash

# if you only need replace images with version set to "main" specify "main-only" argument
REPLACE_SCOPE=$1
# IMG_DIRECTORY can be omitted
REQUIRED_ENV_VARIABLES=('IMG_DIRECTORY' 'IMG_VERSION' 'PROJECT_ROOT')
for VAR in "${REQUIRED_ENV_VARIABLES[@]}"; do
  if [[ -z "${!VAR}" ]]; then
    echo "${VAR} is undefined"
    exit 1
  fi
done

MAIN_ONLY_SELECTOR=""
if [[ ${REPLACE_SCOPE} == "main-only" ]]; then
  MAIN_ONLY_SELECTOR="| select(. == \"*:main$\")"
fi

VALUES_FILE=${PROJECT_ROOT}/config/registry-proxy/values.yaml

if [[ -n "${CONTAINER_REGISTRY}" ]]; then
  yq --inplace ".global.images[] |= sub(\"europe-docker.pkg.dev/kyma-project/\", \"${CONTAINER_REGISTRY}\")" "${VALUES_FILE}"
elif [[ ${PURPOSE} == "local" ]]; then
  echo "Changing container registry to k3d-kyma-registry.localhost:5000"
  # TOOD: why won't it work with k3d registry path?
  yq --inplace '.global.images[] |= sub("europe-docker.pkg.dev/kyma-project/", "")' "${VALUES_FILE}"
fi

IMAGES_SELECTOR=".global.images[] ${MAIN_ONLY_SELECTOR}"
# replace /dev/|/prod/ with /IMG_DIRECTORY/
yq --inplace "(${IMAGES_SELECTOR})|= sub (\"/dev/|/prod/\", \"/${IMG_DIRECTORY}/\") " "${VALUES_FILE}"
# replace the last :.* with :IMG_VERSION, sicne the URL can contain a port number
yq --inplace "(${IMAGES_SELECTOR}) |= sub(\":[^:]+$\",\":${IMG_VERSION}\")" "${VALUES_FILE}"
echo "==== Local Changes ===="
yq '.global.images' "${VALUES_FILE}"
echo "==== End of Local Changes ===="

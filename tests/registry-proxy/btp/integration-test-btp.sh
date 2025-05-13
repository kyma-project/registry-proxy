#!/bin/bash

echo -e "\n--------------------------------------------------------------------------------------\n"
echo "Running kyma integration tests uing connected managed kyma runtime"


echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step1: Generating temporary access for new service account\n"

../../bin/kyma alpha kubeconfig generate --clusterrole cluster-admin --serviceaccount test-sa --output /tmp/kubeconfig.yaml --time 2h

export KUBECONFIG="/tmp/kubeconfig.yaml"
if [[ $(kubectl config view --minify --raw | yq '.users[0].name') != 'test-sa' ]]; then
    exit 1
fi
echo "Running test in user context of: $(kubectl config view --minify --raw | yq '.users[0].name')"

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step2: List modules\n"
../../bin/kyma alpha module list


echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step3: Connecting to a service manager from remote BTP subaccount\n"

# https://help.sap.com/docs/btp/sap-business-technology-platform/namespace-level-mapping?locale=en-US
( cd tf ; curl https://raw.githubusercontent.com/kyma-project/btp-manager/main/hack/create-secret-file.sh | bash -s operator remote-service-manager-credentials )
kubectl create -f tf/btp-access-credentials-secret.yaml || true

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step4: Create service instance reference to a shared object-store service instance\n"

echo "Waiting for CRD btp operator"
while ! kubectl get crd btpoperators.operator.kyma-project.io; do echo "Waiting for CRD btp operator..."; sleep 1; done
kubectl wait --for condition=established crd/btpoperators.operator.kyma-project.io
while ! kubectl get btpoperators.operator.kyma-project.io btpoperator --namespace kyma-system; do echo "Waiting for btpoperator..."; sleep 1; done
kubectl wait --for condition=Ready btpoperators.operator.kyma-project.io/btpoperator -n kyma-system --timeout=180s


# TODO - change after btp operator commands are extracted as btp module cli extension
../../bin/kyma alpha reference-instance \
    --btp-secret-name remote-service-manager-credentials \
    --namespace kyma-system \
    --offering-name objectstore \
    --plan-selector standard \
    --reference-name object-store-reference
kubectl apply -n kyma-system -f ./k8s-resources/object-store-binding.yaml

while ! kubectl get secret object-store-reference-binding --namespace kyma-system; do echo "Waiting for object-store-reference-binding secret..."; sleep 5; done


# Enable Docker Registry
echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step5: Enable Docker Registry from experimental channel (with persistent BTP based storage)\n"
../../bin/kyma alpha module add docker-registry --channel experimental --cr-path k8s-resources/custom-docker-registry.yaml

echo "..waiting for docker registry"
kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/custom-dr -n kyma-system --timeout=360s

sleep 5

dr_external_url=$(../../bin/kyma alpha registry config --externalurl)

# TODO new cli command, for example
# dr_internal_pull_url=$(../../bin/kyma alpha registry config --internalurl)
dr_internal_pull_url=$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system custom-dr -ojsonpath={.status.internalAccess.pullAddress})

../../bin/kyma alpha registry config --output config.json

echo "Docker Registry enabled (URLs: $dr_external_url, $dr_internal_pull_url)"
echo "config.json for docker CLI access generated"

echo -e "\n--------------------------------------------------------------------------------------\n"
echo -e "Step6: Map SAP Hana DB instance with Kyma runtime\n"

../../bin/kyma alpha hana map --credentials-path tf/hana-admin-creds.json

make -C ../../ run-integration

exit 1
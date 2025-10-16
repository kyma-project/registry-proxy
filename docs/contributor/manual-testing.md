# Manual test for Registry Proxy

## Prerequisites

- BTP Kyma Cluster with connectivity-proxy module enabled
- Docker installed and running
- kubectl
- [Registry Proxy repository](https://github.com/kyma-project/registry-proxy)
- [Cloud Connector](https://tools.hana.ondemand.com/#cloud)
- Htpasswd, openssl
- [SapMachine 21 JDK](https://sapmachine.io/)
- Export variables listed below

```bash
export KUBECONFIG={PATH_TO_YOUR_KYMA_KUBECONFIG}
export EMAIL={YOUR_EMAIL}
export NAMESPACE={NAMESPACE_WHERE_WORKLOAD_IS_DEPLOYED}
export CLUSTER_DOMAIN=$(kubectl get cm -n kube-system shoot-info -ojsonpath='{.data.domain}')
export REG_USER_NAME={REGISTRY_USERNAME}
export REG_USER_PASSWD={REGISTRY_PASSWORD}
export IMAGE_NAME="myregistry.kyma:25002/on-prem-nginx:$(date +%F-%H-%M)"
```

## Steps

To learn how to configure Connectivity Proxy and On-Prem Docker registry, see the [Create Kyma Registry Proxy Connection and a Target Deployment](../user/tutorials/01-10-registry-proxy-connection.md) tutorial.

### Install Registry Proxy

If you want to test the changes not yet deployed to the release channel, install the Registry Proxy module manually instead of installing it as a Kyma module.

Inside the `registry-proxy` repository:

1. Create a namespace for the Registry Proxy.

```bash
kubectl create namespace ${NAMESPACE}
```

2. You can get images for your changes by pushing them and grabbing them from the build job or build the Docker images yourself for Controller and Registry Proxy and push them.

```bash
export CONTAINER_REGISTRY="target-registry.example.com"
export TAG="main"
docker buildx build --push --platform linux/amd64 -t ${CONTAINER_REGISTRY}/registry-proxy-operator:${TAG} . -f ./components/operator/Dockerfile --build-arg=PURPOSE="dev" --build-arg=IMG_DIRECTORY="" --build-arg=IMG_VERSION="${TAG}" --build-arg=CONTAINER_REGISTRY="${CONTAINER_REGISTRY}"
docker buildx build --push --platform linux/amd64 -t ${CONTAINER_REGISTRY}/registry-proxy-controller:${TAG} . -f ./components/registry-proxy/Dockerfile
docker buildx build --push --platform linux/amd64 -t ${CONTAINER_REGISTRY}/registry-proxy-connection:${TAG} . -f ./components/connection/Dockerfile
```

3. Install the operator.

```bash
helm install registry-proxy-operator config/operator -n ${NAMESPACE} \
  --set controllerManager.container.image="${CONTAINER_REGISTRY}/registry-proxy-operator:${TAG}"
```

4. Apply the default Registry Proxy CR in the default namespace.

```bash
cat << EOF | kubectl apply -f -
apiVersion: operator.kyma-project.io/v1alpha1
kind: RegistryProxy
metadata:
  name: default
  namespace: ${NAMESPACE}
spec: {}
EOF
```

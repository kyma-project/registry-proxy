---
parser: v2
auto_validation: true
time: 30
primary_tag: software-product>sap-btp\, kyma-runtime
tags: [tutorial>advanced, software-product>sap-business-technology-platform]
author_name: Piotr Halama, Andrzej Pankowski, Filip Rudy
keywords: kyma
---

# Create Kyma Registry Proxy Connection and a Target Deployment

<!-- description --> In this tutorial, you will set up the on-premise Docker Registry and securely download images on your Kyma cluster.

## You will learn

- How to set up Cloud Connector, on-premise Docker Registry.
- How to install the Registry Proxy module and configure the connection.
- How to create a target deployment using an image from the on-premise Docker Registry.

## Prerequisites

- SAP BTP, Kyma runtime enabled
- [Connectivity Proxy and Registry Proxy modules](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US) added
- Docker installed and running
- kubectl installed
- [SapMachine 21 JDK](https://sapmachine.io/) installed
- [Cloud Connector installed](https://tools.hana.ondemand.com/#cloud)
- Htpasswd and openssl installed

### Prepare enviroment

1. Export the following environment variables:

```bash
export KUBECONFIG={PATH_TO_YOUR_KYMA_KUBECONFIG}
export EMAIL={YOUR_EMAIL}
export NAMESPACE={NAMESPACE_WHERE_WORKLOAD_IS_DEPLOYED}
export CLUSTER_DOMAIN=$(kubectl get cm -n kube-system shoot-info -ojsonpath='{.data.domain}')
export REG_USER_NAME={REGISTRY_USERNAME}
export REG_USER_PASSWD={REGISTRY_PASSWORD}
export IMAGE_TAG="$(date +%F-%H-%M)"
export IMAGE_NAME="myregistry.kyma:25002/on-prem-nginx:${IMAGE_TAG}"
```

### Setup Cloud Connector

1. Run the `go.sh` script from the Cloud Connector download.

```bash
NO_CHECK=1 ./go.sh
```

> [!NOTE]
> On your first try, you may need to add an exception in your system settings under **Privacy & Security**.

2. Go to the link specified in the output.

```bash
Cloud Connector <version> started on <link to follow>
```

If the link doesn't work, replace the domain with `127.0.0.1`, for example:
  - Cloud Connector outputs `Cloud Connector 2.18.0 started on https://custom.domain:8443 (master)`.
  - Open `https://127.0.0.1:8443` in the browser.
3. Log in with the default credentials.
   - Username: `Administrator`
   - Password: `manage`
     You will be prompted to change the password; note it.
4. In your SAP BTP subaccount, go to **Connectivity -> Cloud Connectors** and choose **Download Authentication Data**.
5. In Cloud Connector, go to Define **Subaccount -> Add Subaccount**.
6. Choose **Next** and select **Configure using authentication data**.
7. Add the file from the previous step, and choose **Next**.

### Set up on-premise Docker Registry

1. Generate a self-signed certificate:

```bash
mkdir -p certs
openssl req \
   -newkey rsa:4096 -nodes -sha256 -keyout certs/domain.key \
   -addext "subjectAltName = DNS:myregistry.kyma" \
   -x509 -days 365 -out certs/domain.crt
```

<!-- TODO: mac specific -->

2. Add the certificate to the system keychain:

```bash
security add-trusted-cert -d -r trustRoot -k ~/Library/Keychains/login.keychain certs/domain.crt
```

3. Generate an authentication file for the local Docker registry:

```bash
mkdir -p secret
htpasswd -Bbn ${REG_USER_NAME} ${REG_USER_PASSWD} > ./secret/htpasswd
```

4. Add the required configuration file:

```bash
mkdir -p config
cat << EOF > config/config.yml
version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /var/lib/registry
auth:
  htpasswd:
    realm: basic-realm
    path: /secret/htpasswd
http:
  addr: 0.0.0.0:443
  tls:
    certificate: /certs/domain.crt
    key: /certs/domain.key
    minimumtls: tls1.2
    ciphersuites:
      - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
      - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
      - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      - TLS_AES_128_GCM_SHA256
      - TLS_CHACHA20_POLY1305_SHA256
      - TLS_AES_256_GCM_SHA384
      - TLS_AES_128_GCM_SHA256
  headers:
    X-Content-Type-Options: [nosniff]
EOF
```

5. Run the local Docker registry:

```bash
docker run -d \
		-v docker-reg-vol:/var/lib/registry \
		-v $(pwd)/certs:/certs \
		-v $(pwd)/config/config.yml:/etc/distribution/config.yml \
		-v $(pwd)/secret/htpasswd:/secret/htpasswd \
		-p 25002:443 \
		--restart=always \
		--name on-prem-docker-registry \
		registry:3.0.0
```

6. Edit the `/etc/hosts` file and add the following line: `127.0.0.1 myregistry.kyma`
7. In Cloud Connector, go to **Configuration** and select the **On-Premises** tab.

8. Select **+** in the **Backend Trust Store** section, and add the generated `domain.crt` file to the allowlist.

### Configure Cloud Connector on-premise connection

1. Go to the **Cloud to On-Premises** section in the Cloud Connector UI and under the **Mapping Virtual to Internal System** click the **+** button.
   Provide the following information:
   - Back-end Type: Non-SAP System
   - Protocol: HTTPS
   - Internal and Virtual Host: myregistry.kyma
   - Internal and Virtual Port: 25002
   - Uncheck **Allow Principal Propagation**
   - Host in Request Header: Use Internal Host
2. Choose **+** in the **Resources of {your registry name}**
3. Add `/` as the URL path, mark the **Active** checkbox, and select `Path and All Sub-Paths`. Choose **Save**.
4. Select the `Check availability of internal host` button to make sure the Cloud Connector was configured properly, and the cluster can access the on-prem Docker registry.
   ![check-availability.png](check-availability.png)

### Configure Registry Proxy

1. Apply a Connection CR:

```bash
kubectl create namespace ${NAMESPACE}
cat << EOF | kubectl apply -f -
apiVersion: registry-proxy.kyma-project.io/v1alpha1
kind: Connection
metadata:
  name: registry-proxy-myregistry
  namespace: ${NAMESPACE}
spec:
  targetHost: "myregistry.kyma:25002"
EOF
```

2. Get the Connection NodePort number:

```bash
export NODE_PORT=$(kubectl get connections.registry-proxy.kyma-project.io -n ${NAMESPACE} registry-proxy-myregistry -o jsonpath={.status.nodePort})
```

### Configure on-premise Docker Registry

1. Authenticate to the local Docker registry to push the test image:

```bash
docker login myregistry.kyma:25002 -u ${REG_USER_NAME} -p ${REG_USER_PASSWD}
```

2. Create and push a test image to the on-premise Docker Registry:

```bash
echo -e "FROM nginx:alpine\nRUN echo \"<h1>Test image created on $(date +%F+%T)</h1>\" > /usr/share/nginx/html/index.html" | docker buildx build --push --platform linux/amd64 -t ${IMAGE_NAME} -
```

3. Create a Secret for authentication with the on-premise Docker registry:

```bash
kubectl -n ${NAMESPACE} create secret docker-registry on-premise-reg \
    --docker-username=${REG_USER_NAME} \
    --docker-password=${REG_USER_PASSWD} \
    --docker-email=${EMAIL} \
    --docker-server=localhost:${NODE_PORT}
```

4. Adjust the image in the Deployment to use the image tag and the node port from the previous steps, and deploy it on the cluster:

```bash
kubectl run test-workload-on-prem-reg -n ${NAMESPACE} --image="localhost:${NODE_PORT}/on-prem-nginx:${IMAGE_TAG}" --port 80 --overrides='{"metadata":{"labels":{"app":"test-workload-on-prem-reg","sidecar.istio.io/inject": "true"}},"spec":{"imagePullSecrets":[{"name": "on-premise-reg"}]}}'
kubectl create service clusterip test-workload-on-prem-reg -n ${NAMESPACE} --tcp=80:80
cat << EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: test-workload-on-prem-reg
  namespace: ${NAMESPACE}
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - test-workload-on-prem-reg
  rules:
    - noAuth: true
      methods: ["GET"]
      path: /{**}
  service:
    name: test-workload-on-prem-reg
    namespace: ${NAMESPACE}
    port: 80
EOF
```

5. Check if the workload was deployed successfully:

```bash
kubectl -n ${NAMESPACE} get pods -l app=test-workload-on-prem-reg
```

6. Access the deployed Nginx image at the `https://test-workload-on-prem-reg.${CLUSTER_DOMAIN}` address:

```bash
 curl https://test-workload-on-prem-reg.${CLUSTER_DOMAIN}
```

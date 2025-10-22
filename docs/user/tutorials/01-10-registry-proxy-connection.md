# Create Kyma Registry Proxy Connection and a Target Deployment

In this tutorial, you will set up a Connection to the on-premise Docker Registry to securely download images to your Kyma cluster.

## You will learn

- How to set up Cloud Connector.
- How to install the Registry Proxy module and configure the connection.
- How to create a target deployment using an image from the on-premise Docker Registry.
- How to set up a Connection to a Docker Registry with the OAuth authorization.

> [!IMPORTANT] 
> For the basic authorization part, this tutorial assumes that you have a running local Docker registry reachable from a local network on your machine at `myregistry.acme:25002` and that you can push and pull images locally. To set up Docker registry, follow [Set up Local Docker Registry for Testing](../../contributor/running-local-docker-registry.md). Remember that this Docker Registry instance is only good for testing purposes. For the production setup, you want to choose a Docker Registry instance that is available within the target on-premise network. 


## Prerequisites


- SAP BTP, Kyma runtime enabled
- [Connectivity Proxy](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US) and [Registry Proxy](https://kyma-project.io/#/community-modules/user/README?id=quick-install) modules added
- Docker Registry instance available within the on-premise network
- kubectl installed
- [Kyma CLI installed](https://github.com/kyma-project/cli/releases/latest)
- [SapMachine 21 JDK](https://sapmachine.io/) or higher installed
- [Cloud Connector installed](https://tools.hana.ondemand.com/#cloud)


### Prepare Environment

<!-- tabs:start -->

#### **Basic Authorization**

Export the following environment variables:

   ```bash
   export KUBECONFIG={PATH_TO_YOUR_KYMA_KUBECONFIG}
   export EMAIL={YOUR_EMAIL}
   export NAMESPACE={NAMESPACE_WHERE_WORKLOAD_IS_DEPLOYED}
   export CLUSTER_DOMAIN=$(kubectl get cm -n kube-system shoot-info -ojsonpath='{.data.domain}')
   export REG_USER_NAME={REGISTRY_USERNAME}
   export REG_USER_PASSWD={REGISTRY_PASSWORD}
   export DOCKER_REGISTRY_HOST="myregistry.acme"
   export DOCKER_REGISTRY_PORT="25002"
   export DOCKER_REGISTRY="${DOCKER_REGISTRY_HOST}:${DOCKER_REGISTRY_PORT}"

   ```

#### **OAuth Authorization**

Export the following environment variables:

   ```bash
   export KUBECONFIG={PATH_TO_YOUR_KYMA_KUBECONFIG}
   export EMAIL={YOUR_EMAIL}
   export NAMESPACE={NAMESPACE_WHERE_WORKLOAD_IS_DEPLOYED}
   export CLUSTER_DOMAIN=$(kubectl get cm -n kube-system shoot-info -ojsonpath='{.data.domain}')
   export REG_USER_NAME={REGISTRY_USERNAME}
   export REG_USER_PASSWD={REGISTRY_PASSWORD}
   export DOCKER_REGISTRY_HOST={EXISTING_DOCKER_REGISTRY_HOST}
   export DOCKER_REGISTRY_PORT={EXISTING_DOCKER_REGISTRY_PORT}
   export DOCKER_REGISTRY="${DOCKER_REGISTRY_HOST}:${DOCKER_REGISTRY_PORT}"
   export AUTHORIZATION_HOST={OAUTH_HOST_WITH_PORT}
   export IMAGE_TAG={TAG_OF_EXISTING_DOCKER_IMAGE}
   export IMAGE_NAME={NAME_OF_EXISTING_DOCKER_IMAGE}
   ```

<!-- tabs:end -->

### Set Up Cloud Connector

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

### Set Up Trust for the On-Premise Docker Registry


1. In Cloud Connector, go to **Configuration** and select the **On-Premises** tab.
2. Select **+** in the **Backend Trust Store** section, and add the Docker Registry and OAuth server certificates (where applicable) to the allowlist. 

> [!IMPORTANT]
> If you are using the local Docker Registry, as explained in [Set up Local Docker Registry for Testing](../../contributor/running-local-docker-registry.md), add the generated self-signed certificate file (`domain.crt`) to the allowlist. 


### Configure the Cloud Connector On-Premise Connection

<!-- tabs:start -->

#### **Basic Authorization**

1. In the **Cloud to On-Premises** section in the Cloud Connector UI, under the **Mapping Virtual to Internal System** choose **+**, and provide the following information:

   - Back-end Type: Non-SAP System
   - Protocol: HTTPS
   - Internal and Virtual Host: myregistry.acme
   - Internal and Virtual Port: 25002
   - Uncheck **Allow Principal Propagation**
   - Host in Request Header: Use Internal Host

2. Choose **+** in the **Resources of {your registry name}**
3. Add `/` as the URL path, mark the **Active** checkbox, and select `Path and All Sub-Paths`. Choose **Save**.
4. Select the `Check availability of internal host` button to make sure the Cloud Connector was configured properly, and the cluster can access the on-prem Docker registry.

   ![check-availability.png](check-availability.png)


#### **OAuth Authorization**

1. In the **Cloud to On-Premises** section in the Cloud Connector UI, under the **Mapping Virtual to Internal System** choose **+**, and provide the following information:

   - Back-end Type: Non-SAP System
   - Protocol: HTTPS
   - Internal and Virtual Host: {DOCKER_REGISTRY_HOST}
   - Internal and Virtual Port: {DOCKER_REGISTRY_PORT}
   - Uncheck **Allow Principal Propagation**
   - Host in Request Header: Use Internal Host

2. Choose **+** in the **Resources of {your registry name}**
3. Add `/` as the URL path, mark the **Active** checkbox, and select `Path and All Sub-Paths`. Choose **Save**.
4. If the OAuth uses a different host or port than the Docker Registry, create another **Mapping Virtual to Internal System**, and provide the following information:

- Back-end Type: Non-SAP System
- Protocol: HTTPS
- Internal and Virtual Host: {OAUTH_DOMAIN}
- Internal and Virtual Port: {OAUTH_PORT}
- Uncheck **Allow Principal Propagation**
- Host in Request Header: Use Internal Host

5. Choose **+** in the **Resources of {your oauth name}**
6. Add `/` as the URL path, mark the **Active** checkbox, and select `Path and All Sub-Paths`. Choose **Save**.
7. Select the `Check availability of internal host` button to make sure the Cloud Connector was configured properly, and the cluster can access the on-prem Docker registry.

   ![check-availability.png](check-availability.png)

<!-- tabs:end -->

### Configure Registry Proxy

<!-- tabs:start -->

#### **Basic Authorization**

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
     target:
       host: "${DOCKER_REGISTRY}"
   EOF
   ```

#### **OAuth Authorization**

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
     target:
       host: "${DOCKER_REGISTRY}"
       authorization:
         host: "${AUTHORIZATION_HOST}"
   EOF
   ```

<!-- tabs:end -->

2. Get the Connection NodePort number:

   ```bash
   export NODE_PORT=$(kubectl get connections.registry-proxy.kyma-project.io -n ${NAMESPACE} registry-proxy-myregistry -o jsonpath={.status.nodePort})
   ```

### Deploy Container from Image Hosted on the On-Premise Docker Registry 

1. Ensure that the image exists in the target Docker Registry  

   Export environment variables referencing the image, for example:
   ```bash
   export IMAGE_TAG="0.0.1"
   export IMAGE_NAME="on-prem-nginx"
   export IMAGE_PATH="${DOCKER_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
   ```
   
   Authenticate to the target Docker registry to push the test image:

   ```bash
   docker login ${DOCKER_REGISTRY} -u ${REG_USER_NAME} -p ${REG_USER_PASSWD}
   
   echo -e "FROM nginx:alpine\nRUN echo \"<h1>Test image created on $(date +%F+%T)</h1>\" > /usr/share/nginx/html/index.html" | docker buildx build --push --platform linux/amd64 -t ${IMAGE_PATH} -
   ```

2. Create a Secret for authentication with the on-premise Docker registry:

   ```bash
   kubectl -n ${NAMESPACE} create secret docker-registry on-premise-reg \
       --docker-username=${REG_USER_NAME} \
       --docker-password=${REG_USER_PASSWD} \
       --docker-email=${EMAIL} \
       --docker-server=localhost:${NODE_PORT}
   ```

3.  Deploy a container on the cluster:

<!-- tabs:start -->

#### **Kyma CLI**

   ```bash
   kyma app push --name test-workload-on-prem-reg --image "localhost:${NODE_PORT}/${IMAGE_NAME}:${IMAGE_TAG}" --container-port 80 --image-pull-secret on-premise-reg --expose --istio-inject

   Creating deployment default/test-on-prem-nginx3

   Creating service default/test-on-prem-nginx3

   Creating API Rule default/test-on-prem-nginx3

   The test-on-prem-nginx3 app is available under the
   {test-workload-on-prem-reg....}
   ```
#### **kubectl**

   ```bash
   kubectl run test-workload-on-prem-reg -n ${NAMESPACE} --image="localhost:${NODE_PORT}/${IMAGE_NAME}:${IMAGE_TAG}" --port 80 --overrides='{"metadata":{"labels":{"app":"test-workload-on-prem-reg","sidecar.istio.io/inject": "true"}},"spec":{"imagePullSecrets":[{"name": "on-premise-reg"}]}}'
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
<!-- tabs:end -->

3. Check if the workload was deployed successfully:

   ```bash
   kubectl -n ${NAMESPACE} get pods -l app=test-workload-on-prem-reg
   ```

4. Access the deployed Nginx image at the `https://test-workload-on-prem-reg.${CLUSTER_DOMAIN}` address:

   ```bash
    curl https://test-workload-on-prem-reg.${CLUSTER_DOMAIN}
   ```

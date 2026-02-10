# Connection

The `connections.registry-proxy.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data used to manage **Connections** within Kyma. It facilitates establishing a connection to a target container registry through the Connectivity Proxy. 

To get the up-to-date CRD in YAML format, run:

```bash
kubectl get connections.registry-proxy.kyma-project.io -A -o yaml
```

## Sample Custom Resource

The following `Connection` object creates a connection to a target registry through the Connectivity Proxy. **proxyURL** specifies the Connectivity Proxy's URL. **targetHost** defines the target registry's host.

```yaml
apiVersion: registry-proxy.kyma-project.io/v1alpha1
kind: Connection
metadata:
  name: registry-proxy-example
spec:
  proxy:
    url: "http://connectivity-proxy.kyma-system.svc.cluster.local:20003"
  target:
    host: "myregistry.kyma:25002"
  logLevel: debug
```

## Custom Resource Parameters
<!-- TABLE-START -->
### connections.registry-proxy.kyma-project.io/v1alpha1

**Spec:**

| Parameter                               | Type                           | Description                                                                                 |
| --------------------------------------- | ------------------------------ |---------------------------------------------------------------------------------------------|
| **proxy**                               | object                         | Specifies the connection to the proxy. If not set, defaults to the value from the  Registry Proxy CR `.spec.proxy` field.                                                     |
| **proxy.url**                           | string                         | URL of the Connectivity Proxy, with protocol.                                               |
| **proxy.locationID**                    | string                         | Sets the `SAP-Connectivity-SCC-Location_ID` header with given ID on every forwarded request |
| **target** (required)                   | object                         | Specifies the connection to the target registry.                                            |
| **target.host** (required)              | string                         | Specifies the target host.                                                                  |
| **target.authorization**                | object                         | Specifies the authorization method for the connection                                       |
| **target.authorization.host**           | string                         | Name of the host that is used for registry authorization                                    |
| **target.authorization.headerSecret**   | string                         | Name of the secret containing the authorization header to be used for the connection.       |
| **resources**                           | object                         | Defines compute resource requirements for the Connection, such as CPU or memory.            |
| **logLevel**                            | string                         | Sets the desired log level to be used. The default value is `"info"`.                       |
| **nodePort**                            | integer                        | Sets the desired service NodePort number.                                                   |


**Status:**

| Parameter          | Type                           | Description                                                                       |
| ------------------ | ------------------------------ |-----------------------------------------------------------------------------------|
| **nodePort**       | integer                        | Specifies the service NodePort number. Use `localhost:<nodeport>` to pull images. |
| **proxyURL**       | string                         | URL of the Connectivity Proxy.                                                    |
| **conditions**     | \[\]object                     | Specifies an array of conditions describing the status of the Connection.         |

<!-- TABLE-END -->

### Status Reasons

Processing of a `Connection` CR can succeed, continue, or fail for one of these reasons:

| Reason                           | Type                 | Description                                                                                    |
| -------------------------------- | -------------------- | ---------------------------------------------------------------------------------------------- |
| `DeploymentCreated`              | `ConnectionDeployed` | A new Deployment referencing the Connection's configuration was created.                       |
| `DeploymentUpdated`              | `ConnectionDeployed` | The existing Deployment was updated after applying changes to the Connection's configuration.  |
| `DeploymentFailed`               | `ConnectionDeployed` | The Connection's Deployment failed due to an error.                                            |
| `InvalidProxyURL`                | `ConnectionDeployed` | The provided Proxy URL is invalid.                                                             |
| `ConnectionResourcesDeployed`    | `ConnectionReady`    | Resources required for the Connection were successfully deployed.                              |
| `ConnectionResourcesNotReady`    | `ConnectionReady`    | Resources required for the Connection are not ready.                                           |
| `ConnectionEstablished`         | `ConnectionReady`    | The Connection was successfully established.                                                   |
| `ConnectionNotEstablished`      | `ConnectionReady`    | The Connection could not be established.                                                       |
| `ConnectionError`                | `ConnectionReady`    | An error occurred while processing the Connection.                                             |

## Related Resources and Components

These are the resources related to this CR:

| Custom resource                                                                                       | Description                                                                             |
| ----------------------------------------------------------------------------------------------------- |-----------------------------------------------------------------------------------------|
| [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)                   | Manages the Pods required for the Connection functionality.                             |
| [Service](https://kubernetes.io/docs/concepts/services-networking/service/)                           | Exposes the Connection's Deployment as a network service inside the Kubernetes cluster. |

These components use this CR:

| Component             | Description                                                                                                  |
|-----------------------| ------------------------------------------------------------------------------------------------------------ |
| Connection Controller | Manages the lifecycle of the Connection CR and ensures the connection to the target registry is established. |

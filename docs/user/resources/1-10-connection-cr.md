# Connection

The `connections.registry-proxy.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage **Connections** within Kyma. It facilitates establishing a connection to a target container registry through the Connectivity Proxy. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get connections.registry-proxy.kyma-project.io -A -o yaml
```

## Sample Custom Resource

The following `Connection` object creates a connection to a target registry through the Connectivity Proxy. The `proxyURL` specifies the Connectivity Proxy's URL, and the `targetHost` defines the target registry's host.

```yaml
apiVersion: registry-proxy.kyma-project.io/v1alpha1
kind: Connection
metadata:
  name: registry-proxy-example
spec:
  proxyURL: "http://connectivity-proxy.kyma-system.svc.cluster.local:20003"
  targetHost: "myregistry.kyma:25002"
  logLevel: debug
```

## Custom Resource Parameters
<!-- TABLE-START -->
### connections.registry-proxy.kyma-project.io/v1alpha1

**Spec:**

| Parameter                 | Type                           | Description                                                                                 |
|---------------------------| ------------------------------ |---------------------------------------------------------------------------------------------|
| **proxyURL**              | string                         | URL of the Connectivity Proxy, with protocol.                                               |
| **targetHost** (required) | string                         | Specifies the target host.                                                                  |
| **resources**             | object                         | Defines compute resource requirements for the Connection, such as CPU or memory.            |
| **logLevel**              | string                         | Sets the desired log level to be used. The default value is `"info"`.                       |
| **nodePort**              | integer                        | Sets the desired service NodePort number.                                                   |
| **locationID**            | string                         | Sets the `SAP-Connectivity-SCC-Location_ID` header with given ID on every forwarded request |

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
| `ConnectionEstabilished`         | `ConnectionReady`    | The Connection was successfully established.                                                   |
| `ConnectionNotEstabilished`      | `ConnectionReady`    | The Connection could not be established.                                                       |
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

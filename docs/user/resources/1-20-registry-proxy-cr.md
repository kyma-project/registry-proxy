# Registry Proxy

The `registryproxies.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to manage **Registry Proxies** within Kyma. It facilitates the configuration and management of the Registry Proxy service, which acts as a proxy between the Kyma cluster and external container registries. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get registryproxies.operator.kyma-project.io -A -o yaml
```

## Sample Custom Resource

The following `RegistryProxy` object represents a proxy configuration for connecting to a target container registry through the Connectivity Proxy. The `status` field provides information about the state, served status, and conditions of the Registry Proxy.
```yaml
apiVersion: registry-proxy.kyma-project.io/v1alpha1
kind: RegistryProxy
metadata:
  name: registry-proxy-example
spec: {}
```

## Custom Resource Parameters
<!-- TABLE-START -->
### registryproxies.operator.kyma-project.io/v1alpha1

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **state**                | string   | Represents the current state of the Registry Proxy. Possible values: `Ready`, `Processing`, `Error`, `Deleting`, `Warning`. |
| **served**               | string   | Indicates whether the Registry Proxy is actively managed. Possible values: `True`, `False`.     |
| **conditions**           | \[\]object | Specifies an array of conditions describing the status of the Registry Proxy.                  |

<!-- TABLE-END -->

### Status Reasons

Processing of a RegistryProxy CR can succeed, continue, or fail for one of these reasons:

| Reason                        | Type                 | Description                                                                                     |
|-------------------------------| -------------------- | ----------------------------------------------------------------------------------------------- |
| `Configuration`               | `Configured` | The Registry Proxy is being configured.                      |
| `ConfigurationErr`           | `Configured` | An error occurred during the configuration of the Registry Proxy.|
| `Configured`            | `Configured` | The Registry Proxy has been successfully configured.                                          |
| `Installation`             | `Installed` | The Registry Proxy is being installed.                                                        |
| `InstallationErr` | `Installed`    | An error occurred during the installation of the Registry Proxy.                             |
| `Installed` | `Installed`    | The Registry Proxy has been successfully installed.                                          |
| `RegistryProxyDuplicated`      | `Installed`    | A duplicate Registry Proxy was detected.                                                 |
| `Deletion`   | `Deleted`    | The Registry Proxy is being deleted.                                                       |
| `DeletionErr`             | `Deleted`    | An error occurred during the deletion of the Registry Proxy.                                           |
| `Deleted`             | `Deleted`    | The Registry Proxy has been successfully deleted.                                        |
| `ConnectivityProxyUnavailable` | `PrerequisitesSatisfied`    | The Connectivity Proxy StatefulSet status is unknown.                                       |
| `ConnectivityProxyAvailable`   | `PrerequisitesSatisfied`    | The Connectivity Proxy StatefulSet is ready.                                       |

## Related Resources and Components

These are the resources related to this CR:

| Custom resource                                                                                              | Description                                                                                 |
| ----------------------------------------------------------------------------------------------------- |---------------------------------------------------------------------------------------------|
| [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)                   | Manages the Pods required for the Registry Proxy functionality.                             |
| [Service](https://kubernetes.io/docs/concepts/services-networking/service/)                           | Exposes the Registry Proxy's Deployment as a network service inside the Kubernetes cluster. |

These components use this CR:

| Component           | Description                                                                                            |
|---------------------|--------------------------------------------------------------------------------------------------------|
| Registry Proxy Controller | Manages the lifecycle of the Registry Proxy CR and ensures the proxying functionality is operational.  |
| Registry Proxy Service | Acts as a proxy between the Kyma cluster and the target container registry.                            |

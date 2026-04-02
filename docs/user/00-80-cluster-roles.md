# ClusterRoles

Learn about ClusterRoles in the Registry Proxy module.
The Registry Proxy module includes several ClusterRoles that are used to manage permissions for the Registry Proxy operator and to aggregate permissions for end users.

## Registry Proxy Edit ClusterRole

With the `kyma-registry-proxy-edit` ClusterRole, you can edit the Registry Proxy resources. For the available options, see the following table:

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | registryproxies | create, delete, get, list, patch, update, watch |
| operator.kyma-project.io | registryproxies/status | get |

## Registry Proxy View ClusterRole

With the `kyma-registry-proxy-view` ClusterRole, you can view the Registry Proxy resources. For the available options, see the following table:

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | registryproxies | get, list, watch |
| operator.kyma-project.io | registryproxies/status | get |

## Connection Edit ClusterRole

With the `kyma-connection-edit` ClusterRole, you can edit the Connection resources.

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | connections | create, delete, get, list, patch, update, watch |
| operator.kyma-project.io | connections/status | get |

## Connection View ClusterRole

With the `kyma-connection-view` ClusterRole, you can view the Connection resources. For the available options, see the following table:

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | connections | get, list, watch |
| operator.kyma-project.io | connections/status | get |

## Role Aggregation

The Registry Proxy module uses the Kubernetes [role aggregation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to automatically extend the default `edit` and `view` ClusterRoles with Registry Proxy-specific permissions.

- **kyma-registry-proxy-edit**: Aggregated to `edit` ClusterRole
- **kyma-registry-proxy-view**: Aggregated to `view` ClusterRole
- **kyma-connection-edit**: Aggregated to `edit` ClusterRole
- **kyma-connection-view**: Aggregated to `view` ClusterRole

This means that if you have the default Kubernetes `edit` or `view` ClusterRoles, you automatically receive the corresponding Registry Proxy permissions without requiring additional role bindings.


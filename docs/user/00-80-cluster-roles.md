# Cluster Roles

The Registry Proxy module includes several ClusterRoles that are used to manage permissions for the Registry Proxy operator and to aggregate permissions for end users. This document describes all ClusterRoles bundled with the Registry Proxy module.

## Registry Proxy Edit ClusterRole

The `kyma-registry-proxy-edit` ClusterRole allows users to edit Registry Proxy resources.

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | registryproxies | create, delete, get, list, patch, update, watch |
| operator.kyma-project.io | registryproxies/status | get |

## Registry Proxy View ClusterRole

The `kyma-registry-proxy-view` ClusterRole allows users to view Registry Proxy resources.

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | registryproxies | get, list, watch |
| operator.kyma-project.io | registryproxies/status | get |

## Connection Edit ClusterRole

The `kyma-connection-edit` ClusterRole allows users to edit Connection resources.

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | connections | create, delete, get, list, patch, update, watch |
| operator.kyma-project.io | connections/status | get |

## Connection View ClusterRole

The `kyma-connection-view` ClusterRole allows users to view Connection resources.

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| operator.kyma-project.io | connections | get, list, watch |
| operator.kyma-project.io | connections/status | get |

## Role Aggregation

The Registry Proxy module uses Kubernetes [role aggregation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles) to automatically extend the default `edit` and `view` ClusterRoles with Registry Proxy-specific permissions.

- **kyma-registry-proxy-edit**: Aggregated to `edit` ClusterRole
- **kyma-registry-proxy-view**: Aggregated to `view` ClusterRole
- **kyma-connection-edit**: Aggregated to `edit` ClusterRole
- **kyma-connection-view**: Aggregated to `view` ClusterRole

This means that users who are granted the default Kubernetes `edit` or `view` ClusterRoles automatically receive the corresponding Registry Proxy permissions without requiring additional role bindings.


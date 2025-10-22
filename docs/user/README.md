# Registry Proxy Module

> [!IMPORTANT] 
> The Registry Proxy module is a community module and is not automatically updated or maintained. For more information on community modules and how to install them, see [Community Modules](https://kyma-project.io/#/community-modules/user/README.md).


Learn more about the Registry Proxy module. Use it to enable the creation of Kubernetes workloads from images pulled from your on-premises Docker registries. 

The Registry Proxy module uses the Connectivity Proxy service, which leverages [SAP Cloud Connector](https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/cloud-connector) technology.

> [!NOTE] 
> For Connectivity Proxy to work, you must have the SAP Cloud Connector service enabled and configured in the same SAP BTP subaccount as your Kyma cluster, and you must have the SAP Cloud Connector software running on the on-premises part of your setup.

## What Is Registry Proxy?

The Registry Proxy module helps ensure security compliance for organizations with strict security policies, where exposing the Docker registry to the public internet is not acceptable. 

With the Registry Proxy module, you can set up a managed connection between the Kubernetes image puller (containerd) and the Connectivity Proxy service inside your Kyma cluster. Combined with a properly configured SAP Cloud Connector targeting your on-premises Docker registry, this allows you to run workloads in your Kyma cluster using container images hosted on your own infrastructure without exposing your internal registry directly to the internet.


## Features

The Registry Proxy module provides the following features:

- Managed in-cluster infrastructure enabling container runtime in Kyma cluster to connect to an on-premise Docker registry
- Compatibility with Docker Registry HTTP API v2
- Built-in probing mechanism that shows the status of the configured Connections

## Architecture

![Registry Proxy module diagram](../assets/registry-proxy.drawio.svg)

1. Administrator configures Cloud Connector (Cloud-to-on-prem) between SAP BTP Subaccount and the on-premise Docker registry.
2. Administrator configures the `Connection` custom resource.
3. Application Developer reads the `Connection` status, configures the image pull secret, and schedules a Kubernetes workload from a container image.
4. Kubernetes container runtime pulls image request routes through a managed NodePort and reverse-proxy Pod to finally reach the on-premise Docker registry using the Connectivity Proxy service.

### Registry Proxy Operator

When you add the Registry Proxy module, the Registry Proxy Operator installs it in your cluster. It manages the Registry Proxy lifecycle and communicates its status using the RegistryProxy CR.

## API/Custom Resource Definitions

The API of the Registry Proxy module is based on Kubernetes CustomResourceDefinitions (CRDs), which extend the Kubernetes API. To inspect the specification of the module API, see:

- [Connection CRD](./resources/01-10-connection-cr.md)
- [RegistryProxy CRD](./resources/01-20-registry-proxy-cr.md)

## Security Considerations

To learn how to avoid potential threats when using Registry Proxy Connections, see [Registry Proxy Recommendations](00-10-recommendations.md).


## Related Information

- [Registry Proxy tutorials](tutorials/README.md)
- [Registry Proxy resources](resources/README.md)
- [Registry Proxy technical reference](technical-reference/README.md)

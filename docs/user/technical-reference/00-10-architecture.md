# Registry Proxy Architecture

![Diagram](../../assets/registry-proxy.drawio.svg)

1. Administrator configures Cloud Connector (Cloud-to-on-prem) between SAP BTP subaccount and the on-premise Docker registry.

2. Administrator configures the Connection custom resource.

3. The Application Developer reads the Connection status, configures the image pull secret, and schedules a Kubernetes workload from a container image.

4. The Kubernetes container runtime pulls image request routes through a managed NodePort and reverse-proxy Pod to reach the on-premise Docker registry using the Connectivity Proxy service.
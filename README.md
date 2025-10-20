# Registry Proxy

## Status

[![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/registry-proxy/main?label=registry-proxy-operator)](https://github.com/kyma-project/registry-proxy/commits/main/)

[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/registry-proxy)](https://api.reuse.software/info/github.com/kyma-project/registry-proxy)

## Overview

The Registry Proxy module helps ensure security compliance for organizations with strict security policies, for which exposing an internal Docker registry to the public internet is not acceptable.

With the Registry Proxy module, you can set up a managed connection between kubelet and the Connectivity Proxy service inside your Kyma cluster. Combined with a properly configured SAP Cloud Connector targeting your on-premises Docker registry, you can run workloads in your Kyma cluster using container images hosted on your own infrastructure without exposing your internal registry directly to the internet.

## Install


Download the `registry-proxy-operator.yaml` and `default-registry-proxy-cr.yaml` manifests from the [latest](https://github.com/kyma-project/registry-proxy/releases/latest) release.
Apply `registry-proxy-operator.yaml` to install Registry Proxy Operator:

```sh
kubectl apply -f registry-proxy-operator.yaml
```

To get Registry Proxy installed, apply the sample Registry Proxy CR:

```bash
kubectl apply -f default-registry-proxy-cr.yaml
```

## Getting Started

### Prerequisites

- Go version v1.22.0+
- Docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

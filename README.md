# Registry Proxy

Enables kyma users to setup a managed connection between kubelet and an onprem docker registry

## Description

This module allows users to download images from on-prem location through the Connectivity Proxy module.

![Diagram](docs/assets/registry-proxy.drawio.svg)

## Install

Ensure that the `kyma-system` namespace exists:

```sh
kubectl create namespace kyma-system | true
```

Download the `registry-proxy-operator.yaml` and `default-registry-proxy-cr.yaml` manifests from the [latest](https://github.tools.sap/kyma/registry-proxy/releases/latest) release.
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

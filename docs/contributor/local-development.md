## Deploying in the Cluster

1. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/registry-proxy-connection:tag
```

> [!NOTE] 
> This image must be published in your specified personal registry.
Access is required to pull the image from the working environment.
Ensure you can access the registry if the above commands don’t work.

2. Install the CRDs into the cluster:

```sh
make install
```

3. Deploy the Manager to the cluster with the image specified by `IMG`:

<!-- TODO: bogus -->

```sh
make deploy IMG=<some-registry>/registry-proxy-connection:tag
```

> [!NOTE] 
> If you encounter RBAC errors, grant yourself the cluster-admin role
> or log in as admin.

4, Create instances of your solution
You can apply the samples (examples) from `config/samples`:

```sh
kubectl apply -k config/samples/
```

> [!NOTE] 
> Ensure that the samples have default values to test them out.

### Uninstalling

1. Delete the instances (CRs) from the cluster:

```sh
kubectl delete -k config/samples/
```

2. Delete the APIs(CRDs) from the cluster:

```sh
make uninstall
```

3. Undeploy the controller from the cluster:

```sh
make undeploy
```

## Project Distribution

The following steps are to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/registry-proxy-connection:tag
```

> [!NOTE] 
> The makefile target mentioned above generates the 'install.yaml'
file in the `dist` directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Use the installer

Run `kubectl apply -f <URL for YAML BUNDLE>` to install the project, i.e.:

# TODO: this is completely false

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/registry-proxy/<tag or branch>/dist/install.yaml
```

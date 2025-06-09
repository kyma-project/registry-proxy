# Recommendations

## Exclude NodePort Range From VPC Peering Routes

The Registry Proxy module uses the Kubernetes NodePort service, accessible on ports 30000-32767 from all cluster nodes.
If an attacker compromises the cluster, they can use the NodePort service to potentially gain access to the private registry (provided they have the credentials to access it).
You can mitigate this risk by taking the following precautions:
- Exclude the NodePort range (30000-32767) from VPC Peering Routes.
- Use other access controls such as API Gateway.
# System Pods in Mesh

## Example 

23 system pods out of mesh

## Description

When pods are injected with a sidecar, they are considered "in the mesh". This
means that all mesh config will be enforced.

Pods in namespaces `kube-system`, `kube-public`, `istio-system` are not
automatically injected, so they are not reported. The number of system pods is
reported here.

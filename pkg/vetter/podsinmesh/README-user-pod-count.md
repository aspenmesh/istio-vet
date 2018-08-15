# User Pods in Mesh

## Example 

4 user pods in mesh out of 6

## Description

When pods are injected with a sidecar, they are considered "in the mesh". This
means that all mesh config will be enforced. When a pod is not in the mesh, then
mesh functionality is not implemented. When a client and server are
communicating, and only one is in the mesh, then some of the features are
implemented.

Only user pods are reported. Pods in namespaces `kube-system`,
`kube-public`, `istio-system` are not included.

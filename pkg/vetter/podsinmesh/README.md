# Pods In Mesh

The podsinmesh vetter counts the number of Pods in/out of the mesh which can be
useful for detecting misconfigurations or policy violations.

This vetter counts all user Pods (not in namespaces `kube-system`,
`kube-public`, `istio-system`) and reports the number of Pods in the mesh. Pods
are considered in the mesh if they have the correct initializer annotations with
sidecar proxy injected.

It also counts the system Pods (in namespaces `kube-system`, `kube-public`,
`istio-system`) and reports the number of running system pods which are
exempted from the mesh.

## Example

Running podsinmesh vetter returns the following output:

```shell
Summary: "User pod count"
Message: "INFO: 4 user pods in mesh out of 6"

Summary: "System pod count"
Message: "INFO: 23 system pods out of mesh'
```

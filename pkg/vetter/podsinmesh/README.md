# Pods In Mesh

The `podsinmesh` vetter counts the number of Pods in/out of the mesh which can be
useful for detecting misconfigurations or policy violations.

This vetter counts all user Pods (not in namespaces `kube-system`,
`kube-public`, `istio-system`) and reports the number of Pods in the mesh. Pods
are considered in the mesh if they have the correct initializer annotations and
sidecar proxy injected.

It also counts the system Pods (in namespaces `kube-system`, `kube-public`,
`istio-system`) and reports the number of running system pods which are
exempted from the mesh.

## Notes Generated

- [System pod count](README-system-pod-count.md)
- [User pod count](README-user-pod-count.md)

# Mesh Version

The `meshversion` vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://istio.io/docs/concepts/) components running in the mesh.

When automatic sidecar injection is enabled for pods in the mesh, this vetter
compares the version of Istio to the version of the Istio containers for each
pod in the mesh, then generates notes upon version mismatch.

Version mismatch in various components can lead to unexpected behavior or policy
violations due to incompatibility. It is recommended to upgrade the reported
components to the *Istio version*.

## Notes Generated

- [Mismatched sidecar version](README-sidecar-image-mismatch.md)
- [Mismatched init container version](README-init-image-mismatch.md)

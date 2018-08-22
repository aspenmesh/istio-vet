# Mesh Version

The `meshversion` vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://archive.istio.io/v0.8/docs/concepts/) components running in the mesh.

Vetter `meshversion` considers the version of Istio
[Mixer](https://archive.istio.io/v0.8/docs/concepts/policies-and-telemetry/overview/)
 image specified in the `istio-mixer` deployment as the *Istio version* for the cluster.

It compares the versions of other installed Istio components like
[Pilot](https://archive.istio.io/v0.8/docs/concepts/traffic-management/pilot/) 
with the *Istio version* and generates notes on version mismatch.

It also inspects the version of sidecar proxy deployed in pods in the mesh.
Notes are generated if any version differs from *Istio version*.

Version mismatch in various components can lead to unexpected behavior or policy
violations due to incompatibility. It is recommended to upgrade the reported
components to the *Istio version*.

## Notes Generated

- [Mismatched sidecar version](README-sidecar-image-mismatch.md)
- [Mismatched init container version](README-init-image-mismatch.md)


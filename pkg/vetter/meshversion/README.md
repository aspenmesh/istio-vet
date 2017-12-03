# Mesh Version

The `meshversion` vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://istio.io/docs/concepts/) components running in the mesh.

Vetter `meshversion` considers the version of Istio
[Mixer](https://istio.io/docs/concepts/policy-and-control/mixer.html) image
specified in the `istio-mixer` deployment as the *Istio version* for the cluster.

It compares the versions of other installed Istio components like
[Pilot](https://istio.io/docs/concepts/traffic-management/pilot.html) with the
*Istio version* and generates notes on version mismatch.

It also inspects the version of sidecar proxy deployed in pods in the mesh.
Notes are generated if any version differs from *Istio version*.

Version mismatch in various components can lead to unexpected behavior or policy
violations due to incompatibility. It is recommended to upgrade the reported
components to the *Istio version*.

## Example

### Mismatched sidecar version

If the sidecar proxy image version running in a Pod in mesh is different than
the *Istio version* as described above, following note is generated:

```shell
Summary: "Mismatched sidecar version - myapp-xyz-1234"

Message: "WARNING: The pod myapp-xyz-1234 in namespace default is running with
sidecar proxy version 0.2.9 but your environment is running istio version 0.2.12.
Consider upgrading the sidecar proxy in the pod."
```

### Mismatched Istio component version

If the Istio Pilot image version specified in the `istio-pilot` deployment is
different than the *Istio version* as described above, following note
is generated:

```shell
Summary: "Mismatched istio component versions - istio-pilot"

Message: "WARNING: Istio component istio-pilot is running version 0.2.10 but
your environment is running istio version 0.2.12. Consider upgrading the
component istio-pilot"
```

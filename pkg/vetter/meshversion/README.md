# Mesh Version

The meshversion vetter helps detect mismatched, possibly incompatible versions
of [Istio](https://istio.io/docs/concepts/) components running in the mesh.

Vetter meshversion detects the **official installed version** of Istio by inspecting
the version of Istio
[Mixer](https://istio.io/docs/concepts/policy-and-control/mixer.html) image
specified in the `istio-mixer` deployment.

It compares the versions of other installed Istio components like
[Pilot](https://istio.io/docs/concepts/traffic-management/pilot.html) with the
official version and generates notes on version mismatch.

It also inspects the version of sidecar proxy deployed in the pods in the mesh
and generates notes if they differ from the offical version.

Version mismatch in various components can lead to unexpected behavior or policy
violations due to incompatibility. It is recommended to upgrade the reported
components to the same official Istio version.

## Example

### Mismatched sidecar version
If the sidecar proxy image version running in a Pod in mesh is different than
the Istio Mixer image version specified in the `istio-mixer` deployment
following note is generated:

```shell
Summary: "Mismatched sidecar version - myapp-xyz-1234"

Message: "WARNING: The pod myapp-xyz-1234 in namespace default is running with
sidecar proxy version 0.2.9 but your environment is running istio version 0.2.12.
Consider upgrading the sidecar proxy in the pod."
```

### Mismatched Istio component version
If the Istio Pilot image version specified in the `istio-pilot` deployment is
different than the Istio Mixer image version specified in the `istio-mixer`
deployment following note is generated:

```shell
Summary: "Mismatched istio component versions - istio-pilot"

Message: "WARNING: Istio component istio-pilot is running version 0.2.10 but
your environment is running istio version 0.2.12. Consider upgrading the
component istio-pilot"
```

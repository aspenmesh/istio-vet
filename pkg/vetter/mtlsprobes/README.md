# mTLS Probes

The mtlsprobes vetters identifies if
[mTLS](https://istio.io/docs/tasks/security/mutual-tls.html)
 is enabled in the Istio service mesh along with HTTP or TCP
[probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
(Liveness or Readiness) in any of the Pods in the mesh.

Currently, sidecar proxy deployed in the Pod doesn't identify Probe traffic
coming from the kubelet as authenticated traffic when mTLS is enabled. This
causes probe to fail resulting in Pods being restarted continuously by the
kubelet.

Note that the **Exec** command probes can be used with mTLS enabled in Istio
service mesh.

## Example

If a Pod is deployed in the mesh which has mTLS enabled and a
Liveness or Readiness probe defined in the specification following note is
generated:

```shell
Summary: "mTLS and liveness probe incompatible - myapp-xyz-1234"

Message: "ERROR: The pod myapp-xyz-1234 in namespace default uses
liveness probe which is incompatible with mTLS. Consider disabling the
liveness probe or mTLS."
```

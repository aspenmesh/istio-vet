# mTLS Probes

The `mtlsprobes` vetter verifies if
[mTLS](https://istio.io/docs/tasks/security/mutual-tls/)
is enabled in the Istio service mesh along with HTTP or TCP
[probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
(Liveness or Readiness) in any of the Pods in the mesh.

Currently, sidecar proxy deployed in the Pod doesn't identify Probe traffic
coming from the kubelet as authenticated traffic when mTLS is enabled. This
causes probe to fail resulting in Pods being restarted continuously by the
kubelet.

Note that the **Exec** command probes can be used with mTLS enabled in Istio
service mesh.

## Notes Generated

- [Health probes incompatible with mTLS
  probe](README-mtls-probes-incompatible.md)

# Sidecar Image Mismatch

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` is running with
sidecar proxy image `docker.io/istio/proxy_debug:0.3.0` but your environment is
injecting `docker.io/istio/proxy_debug:0.4.0` for new workloads. Consider
upgrading the sidecar proxy in the pod.

## Description

The service mesh functions by injecting a sidecar proxy container into every
Kubernetes pod.  This sidecar talks to other sidecars and the control plane to
provide service mesh to your application running in the pod.

The `istio-inject` configmap specifies which sidecar container image to inject
into new workloads (like Deployments, DaemonSets, and Jobs) when they are
configured in your cluster.

This warning is generated when a pod is detected that is using a sidecar
container from a different image than what is specified in that configmap (for
new workloads).  If you recreated the same workload from scratch, a different
sidecar container image would be injected.

This image may be missing features, bugfixes or security patches.  It may not
be fully compatible with other sidecars or the control plane.

## Suggested Resolution

Upgrade the sidecar image for these workloads to match the version in the
`istio-inject` configmap, by doing one of the following:

- re-create these workloads again so they are injected with the new sidecar.
- editing the workload (for example, the Deployment)

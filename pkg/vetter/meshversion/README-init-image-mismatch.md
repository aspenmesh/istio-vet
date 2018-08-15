# Init Image Mismatch

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` is running with
istio-init image `docker.io/istio/init:0.3.0` but your environment is
injecting `docker.io/istio/init:0.4.0` for new workloads. Consider
upgrading the istio-init container in the pod.

## Description

The service mesh functions by injecting an istio-init container into every
Kubernetes pod.  This init container sets up the pod so that the application
container will send traffic through the sidecar container.

The `istio-inject` configmap specifies which istio-init container image to
inject into new workloads (like Deployments, DaemonSets, and Jobs) when they
are configured in your cluster.

This warning is generated when a pod is detected that is using an
istio-init container from a different image than what is specified in that
configmap (for new workloads).  If you recreated the same workload from
scratch, a different istio-init container image would be injected.

This image may be missing features, bugfixes or security patches.  It may not
be fully compatible with the control plane.

## Suggested Resolution

Upgrade the istio-init image for these workloads to match the version in the
`istio-inject` configmap, by doing one of the following:

- re-create these workloads again so they are injected with the new istio-init container
- editing the workload (for example, the Deployment)

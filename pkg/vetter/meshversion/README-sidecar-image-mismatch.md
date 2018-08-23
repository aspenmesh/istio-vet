# Sidecar Image Mismatch

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` is running with
sidecar proxy image `docker.io/istio/proxyv2:1.0.0` but your environment is
injecting `docker.io/istio/proxyv2:0.8.0` for new workloads. Consider upgrading
the sidecar proxy in the pod.

## Description

The service mesh functions by injecting a sidecar proxy container into every
Kubernetes pod. Sidecars communicate with each other and with the control plane
to enable mesh features.

Whenever a new pod is created in a namespace where automatic sidecar injection
has been enabled, the injector will modify your pod by adding Istio components.

This warning is generated when a pod is using a `istio-proxy` sidecar image that
is different than what the injector uses. If that pod is deleted, the
replacement pod would be injected with a sidecar matching the image from the
`istio-sidecar-injector` configmap. 

Mismatched images can be problematic for different reasons such as: 
- missing features, bugfixes, or security patches
- not compatible with other sidecars or the control plane


## Suggested Resolution

Re-create this pod so it is injected with a new sidecar matching the version in
the configmap. 

If the pod is managed by a deployment or stateful set, etc., you can delete the
pod and the pod will be recreated with the correct version. Before deleteing a
pod, make sure that deleting it will not affect the state of your workload.

# Istio Vet

The istio-vet tool is a utility to validate the configuration of [Istio](https://istio.io/)
and user applications installed in a [Kubernetes](https://kubernetes.io/)
cluster.

## Introduction

The istio-vet utility helps discover incompatible configuration of user
applications and istio components in a kubernetes cluster. Misconfigurations
might cause unexpected or incorrect service mesh behavior which can be easily
detected and fixed using this tool.

Note that istio-vet only **reads** configuration objects from the kubernetes API server.

### Example

Vetter `meshversion` inspects the version of running istio components and the
sidecar version running in pods in the mesh and reports back the following
on version mismatch:

**Summary**: Mismatched sidecar version - `<pod-id>`

**Message**: WARNING: The pod `<pod-id>` in namespace `<namespace>` is running with sidecar proxy
version `<sidecar-version>` but your environment is running istio version
`<istio-version>`. Consider upgrading the sidecar proxy in the pod.

## Running
The official docker image is `quay.io/aspenmesh/istio-vet:master`

Container image can be deployed in a kubernetes cluster or run locally.

### Local
When run locally kube config for the kubernetes cluster needs to be mounted
inside the container.

```shell
docker run --rm -v $HOME/.kube/config:/root/.kube/config quay.io/aspenmesh/istio-vet:master
```

### In-cluster
The istio-vet container can be deployed as a Job in a kubernetes cluster using
the manifest file in install directory.

```shell
kubectl apply -f install/kubernetes/istio-vet.yaml
```

To inspect the output of the istio-vet, use the following command:

```shell
kubectl -n istio-system logs -l "app=istio-vet" --tail=0
```

Note that the Job would have to be manually run every time to get the latest output
from the istio-vet utility.

Please visit [aspenmesh.io](https://aspenmesh.io/) and sign-up to receive
alerts, insights and analytics from your service mesh.

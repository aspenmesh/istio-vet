# Istio Vet

The istio-vet tool is a utility to validate the configuration of [Istio](https://istio.io/)
and user applications installed in a [Kubernetes](http://kubernetes.io/)
cluster.

## Introduction

The istio-vet utility helps discover incompatible configuration of user
applications and istio components in a kubernetes cluster. Misconfigurations
might cause unexpected or incorrect service mesh behavior which can be easily
detected and fixed using this tool.

Note that istio-vet only **reads** configuration objects from the kubernetes API server.

## Running
The official docker image is `quay.io/aspenmesh/istio-vet`

Container image can be deployed in a kubernetes cluster or run locally.

### Local
When run locally kube config for the kubernetes cluster needs to be mounted
inside the container.

```shell
docker run --rm -v $HOME/.kube/config:/root/.kube/config quay.io/aspenmesh/istio-vet
```

# Istio Vet

[![GoDoc](https://godoc.org/github.com/aspenmesh/istio-vet?status.svg)](https://godoc.org/github.com/aspenmesh/istio-vet)
[![Go Report Card](https://goreportcard.com/badge/github.com/aspenmesh/istio-vet)](https://goreportcard.com/report/github.com/aspenmesh/istio-vet)

The istio-vet tool is a utility to validate the configuration of [Istio](https://Istio.io/)
and user applications installed in a [Kubernetes](https://kubernetes.io/)
cluster.

**This tool works with Istio version 0.7.1 and above.**
<br>
<br>

## Introduction

The istio-vet utility helps discover incompatible configuration of user
applications and Istio components in a kubernetes cluster. Misconfigurations
might cause unexpected or incorrect service mesh behavior which can be easily
detected and fixed using this tool.

The istio-vet tool invokes a list of independent `vetters`. Each `vetter`
performs validation on a subset of configurations and generates `notes`
on any misconfiguration.

Note that istio-vet and vetters only **read** configuration objects from
the kubernetes API server.

#### Example

Vetter `meshversion` inspects the version of running Istio components and the
sidecar version deployed in pods in the mesh. It generates the following
note on any version mismatch:

```shell
Summary: "Mismatched sidecar version - myapp-xyz-1234"

Message: "WARNING: The pod myapp-xyz-1234 in namespace default is running with
sidecar proxy version 0.2.10 but your environment is running Istio
version 0.2.12. Consider upgrading the sidecar proxy in the pod."
```
<br>
<br>

## Running Istio-Vet

You can build and run Istio-Vet from this repo, or use the docker image (locally or from within in a kubernetes cluster).


### Building Istio-Vet

If you want to build Istio-Vet in order to contribute, please see the instructions for [Contributors.](https://github.com/aspenmesh/istio-vet#contributing)

#### Installation

* From within your Go environment, navigate to the directory where you want Istio-Vet to be installed.
* Make sure you have Go installed in your Path:
    ```bash
    export GOPATH=$HOME/gocode
    export PATH=$PATH:$GOPATH/bin
    ```
* Run the go command to get Istio-Vet:
    ```bash
    go get github.com/aspenmesh/istio-vet/cmd/vet
    ```

#### Running Istio-Vet
From anywhere in your Go environment, you should be able to run the following command. The `kube.config` file that you specify will be the cluster which Istio-Vet checks.

  ```bash
  KUBECONFIG=<full-path-to-kube.config> vet
  ```
<br>
<br>

### Using Istio-Vet via Docker

Instructions to run Istio-Vet from our official Docker Image:  `quay.io/aspenmesh/istio-vet:master`
#### Local

When run locally, kube config for the kubernetes cluster needs to be mounted
inside the container.

```shell
docker run --rm -v $HOME/.kube/config:/root/.kube/config quay.io/aspenmesh/istio-vet:master
```

#### In-Cluster

The istio-vet container can be deployed as a Job in a kubernetes cluster using
the manifest file in the install directory.

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

<br>
<br>

## Repository Layout

This repository contains code for the vet tool and supported vetters packages.
It includes:

* **pkg/vet** - This directory contains code for the vet utility which is the
  main binary produced by the repository.

* **pkg/vetters** - This directory contains packages for individual vetters,
  helper utility package and the interface definitions for vetters to implement.
  It includes the following vetters:

  * [meshversion](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/meshversion/README.md) -
    This vetter inspects the version of various installed
    Istio components and generates notes on mismatching versions. It also inspects
    the version of sidecar proxy running in pods in the mesh and compares it
    with the installed Istio version and reports back any version mismatch.

  * [mtlsprobes](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/mtlsprobes/README.md) -
    This vetter inspects if mTLS is enabled in Istio service mesh.
    It generates warnings if Liveness or Readiness probe is configured for a Pod
    in a mTLS enabled mesh.

  * [serviceportprefix](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/serviceportprefix/README.md) -
    This vetter inspects services in the Istio mesh and reports back if any
    service port name definition doesn't include Istio recognized port protocol prefixes.

  * [podsinmesh](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/podsinmesh/README.md) -
    This vetter reports back the number of user pods in/out of
    the mesh. It also reports number of system pods running which are exempted
    from the mesh.

  * [applabel](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/applabel/README.md) -
    This vetter inspects the labels defined for the pods in the mesh and
    generates notes if the label `app` is missing in any pod specification.

  * [serviceassociation](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/serviceassociation/README.md) -
    This vetter generates warning if a pod in the mesh is associated with
    multiple services.

  * [danglingroutedestinationhost](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/danglingroutedestinationhost/README.md) -
    This vetter generates warnings if the route destination host in virtual service resource points to services which don't exists in the cluster.

  * [conflictingvirtualservicehost](https://github.com/aspenmesh/istio-vet/blob/master/pkg/vetter/conflictingvirtualservicehost/README.md) -
    This vetter generates warnings if the same host is defined in multiple
    virtual service resources.

More details about vetters can be found in the individual vetters package
documentation.

<br>
<br>

## Contributing
Individuals or business entities who contribute to this project must have
completed and submitted the [F5® Contributor License Agreement](https://github.com/aspenmesh/cla/raw/master/f5-cla.pdf)
to [cla@aspenmesh.io](mailto:cla@aspenmesh.io) prior to their code submission
being included in this project. Please include your github username in the CLA email.

To build Istio-Vet locally, you will need the following:
```bash
STUFF!
```

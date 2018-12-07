# VirtualServices Define the Same Host

## Example

The VirtualServices vs1, vs2 in namespaces default, default define the
same host, reviews.default.svc.cluster.local. A host name can be defined by only one VirtualService.
Consider updating the VirtualService(s) to have unique hostnames.

## Description

Istio requires that all hostnames defined by VirtualServices in your cluster are
unique. Short hostnames (those that do not contain a '\.') are converted to fully qualified domain names (FQDN) that
include the namespace where the VirtualService resource is defined, so short hostnames are allowed to be repeated so long as
they are defined in separate namespaces. Converting short names to FQDN does not apply to hostnames that include a wildcard '\*' prefix, IP
addresses, or web addresses. These must be unique regardless of the namespace in
which they are defined.

## Conflicting Hostnames Samples

### Sample 1

The FQDNs assigned to the hosts below would be reviews.foo.svc.cluster.local and reviews.bar.svc.cluster.local respectively. This is allowed.

```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs1
    namespace: foo
  spec:
    hosts:
    - reviews
    ...
---
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs2
    namespace: bar
  spec:
    hosts:
    - reviews
    ...
```

### Sample 2

The FQDNs assigned to the hosts in the following example would both be reviews.default.svc.cluster.local.
This is not allowed, and will cause indeterminate routing behavior in your
cluster.

```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs3
    namespace: default
  spec:
    hosts:
    - reviews
    ...
---
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs4
    namespace: default
  spec:
    hosts:
    - reviews
    ...
```

The following note will be generated:

```shell
Summary: "Multiple VirtualServices define the same host -
reviews.default.svc.cluster.local"

Message: "ERROR: The VirtualServices vs3, vs4 in namespaces default, default define the
same host, reviews.default.svc.cluster.local. A host name can be defined by only one VirtualService.
Consider updating the VirtualService(s) to have unique hostnames."
```
See [Suggested Resolution](#suggested-resolution) (1) below for an example of how to fix this by
changing the hostnames to be unique.


### Sample 3

The following sample is also not allowed, as it defines the same web address in
two different VirtualService resources. This will also cause indeterminate
routing behavior in your cluster.

```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs5
    namespace: foo
  spec:
    hosts:
    - google.com
    http:
    - match:
      - uri:
          prefix: /search
      route:
      - destination:
          host: search.foo.svc.cluster.local
---
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs6
    namespace: foo
  spec:
    hosts:
    - google.com
    http:
    - match:
      - uri:
          prefix: /mail
      route:
      - destination:
          host: mail.foo.svc.cluster.local
```

The following note will be generated:

```shell
Summary: "Multiple VirtualServices define the same host - google.com"

Message: "ERROR: The VirtualServices vs5, vs6 in namespaces foo, foo define the
same host, google.com. A host name can be defined by only one VirtualService.
Consider updating the VirtualService(s) to have unique hostnames."
```
See [Suggested Resolution](#suggested-resolution) (2) below for an example of how to fix this by
merging the rules of the two VirtualService resources into one VirtualService
resource.

## Suggested Resolution <a id="suggested-resolution"></a>

You can do one of these two things:

1. **Make the hostnames unique.** Change the hostnames defined in the
   conflicting VirtualServices to be unique. The following VirtualServices have
unique hostnames "reviews" and "ratings", which would resolve the issue for
Sample 2 above.

```yaml
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: vs3
      namespace: default
    spec:
      hosts:
      - reviews
      ...
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: vs4
      namespace: default
    spec:
      hosts:
      - ratings
      ...
```

2. **Merge the conflicting VirtualServices.** Merge the rules defined in the
   conflicting VirtualServices into one VirtualService resource. The following
VirtualService would resolve the issue for Sample 3 above, as the rules are
merged and only a single VirtualService with the "google.com" hostname remains.

```yaml
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: vs5
      namespace: foo
    spec:
      hosts:
      - google.com
      http:
      - match:
        - uri:
            prefix: /search
        route:
        - destination:
            host: search.foo.svc.cluster.local
      - match:
        - uri:
            prefix: /mail
        route:
        - destination:
            host: mail.foo.svc.cluster.local
```

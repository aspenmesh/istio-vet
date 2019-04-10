# Conflicting VirtualService Host

The `conflictingvirtualservicehost` vetter inspects the [Virtual
Service(s)](https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/)
resources in your cluster and generates errors if more than one of them define
the same hostname and gateway. When multiple VirtualServices define the same hostname and gateway, it
can cause indeterminite routing behavior in your cluster.

Istio requires that each VirtualService uses a unique combination of hostname
and gateway. Short hostnames (those that do not contain a '\.') are converted to
fully qualified domain names (FQDN) that include the namespace where the
VirtualService resource is defined, so short hostnames are allowed to be
repeated so long as they are defined in separate namespaces. Converting short
names to FQDN does not apply to hostnames that include a wildcard '\*' prefix,
IP addresses, or web addresses. These must be unique regardless of the namespace
in which they are defined.

It is recommended that you make the hostname and gateway unique per VirtualService, or merge the conflicting
VirtualServices into one VirtualService resource.

## Notes Generated

- [VirtualServices Define the Same Host](README-host-in-multiple-vs.md)

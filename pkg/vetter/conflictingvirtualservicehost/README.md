# Conflicting VirtualService Host

The `conflictingvirtualservicehost` vetter inspects the [Virtual
Service(s)](https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/)
resources in your cluster and generates errors if more than one of them define
the same hostname and the same route. When multiple VirtualServices
define the same hostname, Pilot will try to merge those virtual
services, which can happen in an indeterminite order and cause unexpected
behavior in your cluster. Additionally, if two virtual services define the
same hostname and at least one of them uses sidecar routing (i.e., not
attached to a specific gateway), merging cannot occur.

Istio requires that each VirtualService uses a unique combination of hostname,
 gateway, and route. Short hostnames (those that do not contain a '\.') are
 converted to fully qualified domain names (FQDN) that include the namespace
 where the VirtualService resource is defined, so short hostnames are allowed to
 be repeated so long as they are defined in separate namespaces. Converting short
names to FQDN does not apply to hostnames that include a wildcard '\*' prefix,
IP addresses, or web addresses. These must be unique regardless of the namespace
in which they are defined.

## Notes Generated

- [VirtualServices Define the Same Host](README-host-in-multiple-vs.md)

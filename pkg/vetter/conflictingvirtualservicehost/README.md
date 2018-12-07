# Conflicting VirtualService Host

The `conflictingvirtualservicehost` vetter inspects the [Virtual
Service(s)](https://istio.io/docs/reference/config/istio.networking.v1alpha3/#VirtualService)
resources in your cluster and generates errors if more than one of them define 
the same hostname. When multiple VirtualServices define the same hostname, it 
can cause indeterminite routing behavior in your cluster.

Istio requires that all hostnames defined by VirtualServices in your cluster are
unique. Short hostnames (those that do not contain a '\.') are converted to 
fully qualified domain names (FQDN) that include the namespace where the VirtualService 
resource is defined, so short hostnames are allowed to be repeated so long as
they are defined in separate namespaces. Converting short names to FQDN does not 
apply to hostnames that include a wildcard '\*' prefix, IP addresses, or web 
addresses. These must be unique regardless of the namespace in which they are 
defined.

It is recommended that you make the hostnames unique, or merge the conflicting
VirtualServices into one VirtualService resource.

## Notes Generated

- [VirtualServices Define the Same Host](README-host-in-multiple-vs.md)

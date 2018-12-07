# Dangling Route Destination Host

## Example

The VirtualService vs-a in namespace default has route destination
host(s) svc-a.default.svc.cluster.local pointing to service(s) which don't exist.
Consider adding the services or removing the destination hosts from the
VirtualService resource.

## Description

Any HTTP request routed to destination hosts that do not exist will return a 
`503 Service Unavailable` response code as there is no backend service to 
fulfill the request.

Note that only route destination hosts ending in `.svc.cluster.local` and short
names (host which don't contain any `.`) are inspected by this vetter as
these hosts are implemented by services in the cluster.

## Dangling Route Sample

If any of the following VirtualService resources exist in namespace `default`:

```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs-a
    namespace: default
  spec:
    hosts:
    - mysvc.default.svc.cluster.local
    http:
    - route:
      - destination:
          host: svc-a.default.svc.cluster.local
  ---
  apiVersion: networking.istio.io/v1alpha3
  kind: VirtualService
  metadata:
    name: vs-a
    namespace: default
  spec:
    hosts:
    - mysvc.default.svc.cluster.local
    http:
    - route:
      - destination:
          host: svc-a
```

but the corresponding service `svc-a` is not present in the namespace
`default` the following note is generated:

```shell
Summary: "Dangling route destination - vs-a"

Message: "WARNING: The VirtualService vs-a in namespace default has route destination
host(s) svc-a.default.svc.cluster.local pointing to service(s) which don't exist.
Consider adding the services or removing the destination hosts from the
VirtualService resource."
```
See [Suggested Resolution](#suggested-resolution) below for ways to resolve the 
dangling route destination host.

## Suggested Resolution <a id="suggested-resolution"></a>

- **Create the missing service(s).** Create the service(s) mentioned in the
  VirtualService resource(s).

- **Route to existing services.** Update the VirtualService(s) to route to
  existing services in the cluster. 

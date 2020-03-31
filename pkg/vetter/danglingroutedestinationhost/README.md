# Dangling Route Destination Host

The `danglingroutedestinationhost` vetter inspects the
[VirtualService(s)](https://istio.io/docs/reference/config/networking/virtual-service/)
resources in your cluster and generates warning notes if any of the route
[destination
hosts](https://istio.io/docs/reference/config/networking/virtual-service/#Destination)
point to services which don't exist in the cluster. Any HTTP request routed
to these hosts will return a `503 Service Unavailable` response code as there is
no backend service to fulfill the request.

Note that only route destination hosts ending in `.svc.cluster.local` and short
names (host which don't contain any `.`) are inspected by this vetter as
these hosts are implemented by services in the cluster.

It is recommended to either create the service(s) mentioned in the
VirtualService(s) resources or update the VirtualService(s) to route to existing
services in the cluster.

## Notes Generated

- [Destination host service(s) do not exist](README-dangling-route-destination.md)

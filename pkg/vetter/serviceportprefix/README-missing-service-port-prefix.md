# Missing Prefix in Service

## Example

The service `your-service` in namespace `your-apps` contains port names not
prefixed with mesh supported protocols. Consider updating the service port name
with one of the mesh recognized prefixes.

## Description

The sidecar proxy needs to understand the protocols it is proxying so that it
can apply policies.  Istio Pilot decides which protocol is in use by examining
the service port name, and checking if it is named like `<protocol>-<anything
else>`, for instance `grpc-userapi` (treated as the `grpc` protocol) or
`http2-clients` (treated as the `http2` protocol).

## Suggested Resolution

If the service port is using one of the supported protocols, then you should
add an indicator.  For instance, if your service port named `client` is http2,
rename the service port to `http2-client`.

If the service port is using a layer 7 protocol that the sidecar does not
understand, you should name the service port with a `tcp-` or `udp-` in front
of it to indicate that you are opting out of layer 7 policy from the service
mesh for this service port.  For instance, if your service port named `backend`
is using an unknown protocol that runs on top of tcp, rename the service port
to `tcp-backend`.

In version 1.1.0, these protocols are supported: `grpc`, `http`, `http2`, `https`,
`mongo`, `redis`, `tcp`, `tls`, `udp`.

## See Also

- [Pod and Service Requirements](https://istio.io/docs/setup/kubernetes/prepare/requirements/)

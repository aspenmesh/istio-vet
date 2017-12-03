# Service Port Prefix

The `serviceportprefix` vetter inspects the port names defined in the services in
mesh and generates notes if they are missing Istio recognized port
protocol prefixes.

Service port names need to be prefixed with the recognized
[names](https://istio.io/docs/setup/kubernetes/sidecar-injection.html) for Istio
routing features to work correctly. If a port name doesn't begin with a
recognized prefix or is unnamed, traffic on the port is treated as plain TCP or
UDP depending on the port protocol.

Port names of the form `<protocol>-<suffix>` or `<protocol`> are allowed for
the following protocols:

* http
* http2
* grpc
* mongo
* redis
* tcp

Note that `tcp` protocol prefix can be used to indicate that the port
is for TCP protocol. Service ports with protocol type UDP are also excluded
from this prefix requirement.

## Example

If a Service in mesh is defined with any of the following port definitions:

```shell
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
```
or

```shell
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
    name: myport
```
following note is generated:

```shell
Summary: "Missing prefix in service - myapp-xyz-1234"

Message: "WARNING: The service myapp-service in namespace default
contains port names not prefixed with mesh supported protocols.
Consider updating the service port name with one of the mesh recognized prefixes."
```

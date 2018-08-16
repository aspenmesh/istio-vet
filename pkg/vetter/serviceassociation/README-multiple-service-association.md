# Service Association

## Example

The services `svc-a, svc-b` in namespace `default` are
associated with the pod `myapp-xyz-1234`. Consider updating
the service definitions ensuring the pod belongs to a single service.

## Description

Pods must belong to a single kubernetes service for service mesh to function
correctly.


## Suggested Resolution

Update the services mentioned in the note so that only one is associated with any given pod.

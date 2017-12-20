# Service Association

The `serviceassociation` vetter inspects the services associated with pods in
the mesh and generates notes if a pod is associated with multiple services.
Pods must belong to a single kubernetes service for service mesh to function
correctly.

It is recommended to update the services mentioned in the generated
notes so that the pods are only associated with a single service.

## Example

If the services `svc-a` and `svc-b` are both associated with the pod
`myapp-xyz-1234` following note is generated:

```shell
Summary: "Multiple service association - svc-a, svc-b"

Message: "ERROR: The services svc-a, svc-b in namespace default are
associated with the pod myapp-xyz-1234. Consider updating
the service definitions ensuring the pod belongs to a single service."
```


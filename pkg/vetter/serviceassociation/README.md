# Service Association

The `serviceassociation` vetter inspects the services associated with pods in
the mesh and generates notes if a pod is associated with multiple services.
Pods must belong to a single kubernetes service for service mesh to function
correctly.

It is recommended to update the services mentioned in the generated
notes so that the pods are only associated with a single service.

## Notes Generated

- [Multiple service associations](README-multiple-service-association.md)

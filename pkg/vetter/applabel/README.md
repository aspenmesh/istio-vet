# App Label

The `applabel` vetter inspects the labels defined for the pods in the mesh and
generates notes if the label `app` is missing on any pod.

The label `app` is used to add contextual information in tracing information
collected by the mesh. It is recommended to add uniqiue and meaningful `app`
label to the pods in the mesh in order to collect useful tracing data.

## Example

Following note is generated if `app` label is missing:

```shell
Summary: "Missing app label - myapp-xyz-1234"

Message: "WARNING: The pod myapp-xyz-1234 in namespace default is
missing "app" label. Consider adding the label "app" to the pod."
```

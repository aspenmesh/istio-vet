# App Label

The `applabel` vetter inspects the labels defined for the pods in the mesh and
generates notes if the label `app` is missing on any pod.

The label `app` is used to add contextual information in tracing information
collected by the mesh. It is recommended to add unique and meaningful `app`
label to the pods in the mesh in order to collect useful tracing data.

## Notes Generated

- [Missing app label](README-missing-app-label.md)


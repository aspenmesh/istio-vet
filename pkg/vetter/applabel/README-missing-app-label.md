# Missing App Label

## Example 

The pod `myapp-xyz-1234` in namespace `default` is missing "app" label.
Consider adding the label "app" to the pod.

## Description

The label `app` is used to add contextual information in tracing information
collected by the mesh.

## Suggested Resolution

Add a unique and meaningful `app` label to the pod in order to collect useful
tracing data.


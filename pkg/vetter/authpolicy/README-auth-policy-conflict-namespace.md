# Conflicting Authorization Policies for Namespaces

## Example

Multiple authentication policies `pol-1, pol-2` in namespace `ns-1`
set the namespace-wide config which will cause unwanted behavior. Update
policies to remove conflicts.

## Description

Authentication policies are used to define the authentication and mTLS
requirements that a workload requires to accept traffic. Authentication policies
can target an entire namespace, a service in that namespace, or a port of that
service. Policies for more specific targets override less specific policies. 

When two policies have the same target, the behavior is indeterminate; at any
time you may observe the behavior specified in one policy or another.


## Policy Conflicts: Namespaces

If an authentication policy doesn't list any targets, it is targeting the entire
namespace. The "Conflicting Authorization Policies for Namespace" error
indicates that you have more than one policy that affects the namepspace-wide
policy, which will cause indeterminate behavior.


### Sample Authentication Policies for Namespaces

In this sample, there is a conflict. There are two policies that apply to the `ns-1` namespace. Neither specifies a target, so they are both targeting
the namespace-wide policy. The two policies conflict, and you will get unwanted
behavior - sometimes the first policy will apply, and sometimes the second will
apply.

```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-1
    namespace: ns-1
spec:
    peers:
    - mls: {}
```

```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-2
    namespace: ns-1
spec:
    peers: 
    # empty specifies that mTLS is off
```

In this sample, there are no conflicts. `pol-1` turns mTLS on for
all services in namespace `ns-2`, and `pol-2` turns it off for
service `svc-1` in the same namespace. The two policies have overlapping targets,
but the second policy is more specific than the first, so there is no conflict.
The second policy will be applied only to the workload for service `svc-1`, while
the specifications of the policy without a named target will remain in effect
for all other targets.

```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-1
    namespace: ns-2
spec:
    peers: 
    - mtls: {} 
```


```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-2
    namespace: ns-2
spec:
    targets: 
    - name: svc-1
    peers: 
    # empty specifies that mTLS is off
 
```


## Suggested Resolution

The error message lists the conflicting policies. Locate the affected policies
and compare the settings, then remove the conflict/s, perhaps by adjusting the specificity of the target/s. 



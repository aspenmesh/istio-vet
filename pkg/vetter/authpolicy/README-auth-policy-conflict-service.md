# Conflicting Authorization Policies for Services 

## Example

Multiple authentication policies `pol-1, pol-2` in namespace `ns-1`
set the service-wide config for `svc-1` which will cause unwanted behavior.
Update policies to remove conflicts.

## Description

Authentication policies are used to define the authentication and mTLS
requirements that a workload requires to accept traffic. Authentication policies
can target an entire namespace, a service in that namespace, or a port of that
service. Policies for more specific targets override less specific policies. 

When two policies have the same target, the behavior is indeterminate; at any
time you may observe the behavior specified in one policy or another.

## Policy Conflicts: Target Services

If an authentication policy lists a target service, but does not list ports for
that service, it is targeting the whole service. The "Conflicting Authorization
Policies for Services" error indicates that you have more than one policy that
affects a service-wide policy, which will cause indeterminate behavior.


### Sample Authentication Policies for Target Services

In this sample, there is a conflict. There are two policies for the service
`svc-1` in namespace `ns-1`. The first policy turns mTLS on for `svc-1` and
`svc-2`, but the second policy turns mTLS off for `svc-1`. The two
policies conflict, and you will get unwanted behavior for `svc-1` - sometimes
the first policy will apply, and sometimes the second will apply.

```yaml
apiVersion: v1alpha1 
kind: Policy 
metadata: 
    name: pol-1
    namespace: ns-1
spec:   
    targets: 
    - name: svc-1
    - name: svc-2
    peers: 
        -mtls: {}

```

```yaml
apiVersion: v1alpha1 
kind: Policy 
metadata: 
    name: pol-2
    namespace: ns-1 
spec:   
    targets: 
    - name: svc-1
    peers: 
    # empty specifies that mTLS is off
        
```

In this sample, there are no conflicts. `pol-1` turns mTLS on for
all services in namespace `ns-2`, while `pol-2` applies only to the
service `svc-1` in the namespace `ns-2`. The two policies have overlapping targets
due to the namespace-wide scope of the first, but the second policy is more
specific than the first so there is no conflict. The second policy will be
applied only to the workload for service `svc-1`, while the policy without a named
target will remain in effect for all other targets.

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


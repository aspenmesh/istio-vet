# Conflicting Authorization Policies for Ports 

## Example

Multiple authentication policies `pol-1, pol-2` in namespace `ns-1`
sets the service port config for `svc-1: 8000` which will cause unwanted
behavior. Update policies to remove conflicts.

## Description

Authentication policies are used to define the authentication and mTLS
requirements that a workload requires to accept traffic. Authentication policies
can target an entire namespace, a service in that namespace, or a port of that
service. Policies for more specific targets override less specific policies. 

When two policies have the same target, the behavior is indeterminate; at any
time you may observe the behavior specified in one policy or another.

## Policy Conflicts: Ports

If an authentication policy lists a port for a target service, it targets only the workload for that port of the targeted service. The "Conflicting Authorization Policies for Ports" error indicates that you have more than one policy that
affects the port for the same service, which will cause indeterminate behavior.


### Sample Authentication Policies:

In this sample, there is a conflict. There are two policies that target the same port (8001) for service `svc-1` in the namespace `ns-1`. The first policy turns mTLS on for ports 8000 and 8001 while the second policy turns mTLS off for port 8001. The two policies conflict, and you will get unwanted behavior for traffic to port 8001 - sometimes the first policy will apply, and sometimes the second will apply.

```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-1
    namespace: ns-1
spec:
    targets: 
    - name: svc-1
        ports:
        - number: 8000
        - number: 8001
    peers: 
        - mtls: {}
 
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
        ports:
        - number: 8001
    peers: 
    # empty specifies that mTLS is off
```

In this sample, there are no conflicts. The first policy turns mTLS on for `svc-1` in namespace `ns-2`, while the second policy turns mTLS off only for port 8000 of the same service. The two policies have overlapping targets due to the scope of the first, but the second policy is more specific so there is no conflict. The second policy will be applied only to `svc-1` workloads for port 8000, while the policy without a port listed will remain in effect for all other ports of that service.

```yaml
apiVersion: v1alpha1
kind: Policy
metadata:
    name: pol-1
    namespace: ns-2
targets: 
    - name: svc-1
    peers: 
        -mtls: {} 
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
        ports:
        - number: 8000
    peers: 
    # empty specifies that mTLS is off
 
```

## Suggested Resolution

The error message lists the conflicting policies. Locate the affected policies
and compare the settings, then remove the conflict/s, perhaps by adjusting the specificity of the target/s.

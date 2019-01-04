# Auth Policy Conflict

Authentication policies are used to define the authentication and mTLS
requirements that a workload requires to accept traffic. Authentication policies
can target an entire namespace, a service in that namespace, or a port of that
service. Policies for more specific targets override less specific policies. 

When two policies have the same target, the behavior is indeterminate; at any
time you may observe the behavior specified in one policy or another.

## Notes Generated

- [Conflicting Authorization Policies for Ports](README-auth-policy-conflict-port.md)
- [Conflicting Authorization Policies for
  Services](README-auth-policy-conflict-service.md)
- [Conflicting Authorization Policies for
  Namespaces](README-auth-policy-conflict-namespace.md)


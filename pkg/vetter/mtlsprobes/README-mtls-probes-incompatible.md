# mTLS and Liveness Probes Incompatible

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` uses liveness probe
which is incompatible with mTLS.  Consider disabling the liveness probe or
mTLS.

## Description

mTLS is a feature where all pod-to-pod traffic in the service mesh is encrypted
with mutual TLS.  In this context, mutual means that the client sidecar
authenticates the server sidecar and the server sidecar authenticates the
client sidecar at the beginning of the connection before sending encrypted
traffic.

For sidecars in the mesh, the Istio CA takes care of giving them keys and certs
that will be trusted by other sidecars.  However, the liveness probe or
readiness probe involves the kubelet connecting directly to the sidecar to
probe if the application is alive or ready.  The kubelet does not have a key
and cert that the sidecar would trust, so the sidecar cannot authenticate the
client kubelet and rejects the connection.

## Suggested Resolution

You can do one of these things:

- **Use a Liveness Command Instead.** As a workaround, you can use a liveness
  or readiness command.  Instead of the kubelet connecting to the pod over the
local network, the kubelet will start another process in the pod and that
process runs to decide if the pod is healthy or not.  The command must be
already available in the container you choose in the pod (as an example, if you
specify the curl command below as a liveness command probe, but curl is not
installed in your pod, the liveness command will fail).  Since the command is
being run from inside the pod, it does not use the mesh and is not subject to
mTLS.

    ```yaml
    livenessProbe:
      exec:
        command:
        - curl
        - -f
        - http://localhost:8080/healthz
      initialDelaySeconds: 10
      periodSeconds: 5
    ```

- **Disable mTLS.**  Pod-to-pod traffic is still encrypted on the wire but it
  is possible that a rogue, non-mesh pod could impersonate another mesh pod
when doing pod-to-pod communication.

- **Disable liveness probes.** You can remove liveness and readiness probes
  from your pod spec.  In this case, the kubelet will not be able to determine
when your pod is ready to enter service, and if it has become unhealthy.  So,
as long as your pod is running, it will be considered healthy.  If you choose
this technique, make sure your pod will exit if it becomes unhealthy.

- **Upgrade Istio.** This issue will be fixed in a future version of istio but
  is not yet fixed in any released version.

## See also

- [Istio mTLS](https://istio.io/docs/concepts/security/)
- [Liveness Commands](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-a-liveness-command)
- [istio/old_auth_repo#262](https://github.com/istio/old_auth_repo/issues/262)
- [istio/old_auth_repo#292](https://github.com/istio/old_auth_repo/issues/292)
 

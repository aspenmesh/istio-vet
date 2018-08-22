# Liveness or Readiness Probe Incompatible with mTLS

## Example

The pod `your-app-45574414-qhgq3` in namespace `your-app` uses liveness probe
which is incompatible with mTLS.  Consider disabling the liveness probe or
mTLS, or disabling mTLS for the port of the liveness probe.

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
client kubelet and rejects the connection.  If mTLS is enabled for the
liveness/readiness probe, the Pod will re-start continually and eventually enter
a CrashLoopBackOff state.

## Auth Policies Enabling/Disabling mTLS

Authorization policies can enable or disable mTLS at the port, service name, and
service namespace level. This will override the global mTLS setting for the
port/name/namespace specified.

### Sample Configurations and Auth Policies

#### Sample 1
 
The following configuration, which defines a liveness probe in the deployment
and also has mTLS globally enabled, but no authentication policy for service
that is 
associated with the Pod on that port would cause the Pod to re-start and enter a
CrashLoopBackOff state:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  labels:
    app: httpbin
spec:
  ports:
  - name: http
    port: 8000
  selector:
    app: httpbin
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: httpbin
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: httpbin
        version: v1
    spec:
      containers:
      - image: docker.io/citizenstig/httpbin
      imagePullPolicy: IfNotPresent
      name: httpbin
      ports:
        - containerPort: 8000
      livenessProbe:
        httpGet:
          path: /status/200
          port: 8000
        initialDelaySeconds: 5
        periodSeconds: 5
```
From here, running `kubectl get pods -l app=httpbin` would produce output
similar to the following:

```shell
NAME                       READY     STATUS             RESTARTS   AGE
httpbin-1a23bc456d-7ef8g   1/2       CrashLoopBackOff   3          1m
```
If the Pod is deployed in the mesh with the above configuration, the following
note will be generated:

```shell
Summary: "mTLS and liveness probe incompatible - httpbin-1a23bc456d-7ef8g"

Message: "ERROR: The pod httpbin-1a23bc456d-7ef8g in namespace default uses
liveness probe which is incompatible with mTLS. Consider disabling the
liveness probe or mTLS, or disabling mTLS for the port of the liveness probe."
```
See [Suggested Resolution](#suggested-resolution) below for an example of how to 
fix this by defining an authentication policy for your service.

#### Sample 2

Using the same deployment configuration as above, which defines a liveness probe 
in it, but instead having global mTLS **disabled** and an authentication policy
for the service that is associated with the Pod on that port will also cause 
the Pod to re-start and enter a CrashLoopBackOff state.

An example of an authentication policy that enables mTLS for the same port as
the liveness probe:

```yaml
apiVersion: authentication.istio.io/v1alpha1
kind: Policy
metadata:
  name: httpbin
  namespace: default
spec:
    targets:
    - name: httpbin
      ports:
      - number: 8000
    peers:
    - mtls:
```
Again, running `kubectl get pods -l app=httpbin` would produce output
similar to the following:

```shell
NAME                       READY     STATUS             RESTARTS   AGE
httpbin-1a23bc456d-7ef8g   1/2       CrashLoopBackOff   3          1m
```

If the Pod is deployed in the mesh with the above configuration, the same note
will be generated as the note above in Sample 1.

See [Suggested Resolution](#suggested-resolution) below for an example of how to 
fix this by defining an authentication policy for your service.


## Suggested Resolution <a id="suggested-resolution"></a>

You can do one of these things:

- **Define an Authentication Policy for your Service.** As of Istio v0.8 you can
  define an authentication policy that disables mTLS for the port of the 
liveness/readiness probe.  It is recommended to use a port that is specific for
health probes, and to disable mTLS on that port.  If the application port is
also used for the health port and mTLS is disabled on this shared port, mTLS
will be disabled for the application port. By disabling mTLS for a health-specific
port, this will override the global mTLS setting only for the
port of the liveness/readiness probe.  The following authentication policy would
disable mTLS for the probe port in both Sample 1 and Sample 2 above.

    ```yaml
    apiVersion: "authentication.istio.io/v1alpha1"
    kind: "Policy"
    metadata:
      name: "httpbin"
      namespace: "default"
    spec:
        targets:
        - name: httpbin
          ports:
          - number: 8000
        peers:
    ```

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

- [Istio mTLS](https://archive.istio.io/v0.8/docs/concepts/security/mutual-tls/)
- [Liveness Commands](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-a-liveness-command)
- [istio/old_auth_repo#262](https://github.com/istio/old_auth_repo/issues/262)
- [istio/old_auth_repo#292](https://github.com/istio/old_auth_repo/issues/292)
 

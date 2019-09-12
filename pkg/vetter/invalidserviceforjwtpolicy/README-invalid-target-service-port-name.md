# Invalid Service For JWT Authentication Policy

## Example

The JWT Authentication Policy `platform-jwt-authpolicy` in namespace `platform` targets service `platform-api` in the same
namespace. The target services port names are not prefixed with mesh supported protocols.

```yaml
apiVersion: "authentication.istio.io/v1alpha1"
kind: "Policy"
metadata:
  name: platform-jwt-authpolicy
  namespace: platform
spec:
  targets:
  - name: platform-api
  origins:
  - jwt:
      issuer: "testing@secure.istio.io"
      jwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.2/security/tools/jwt/samples/jwks.json"
  principalBinding: USE_ORIGIN

```

```yaml
apiVersion: v1
kind: Service
metadata:
  name: platform-api
  namespace: platform
spec:
  selector:
    app: platform-api
  ports:
    - protocol: TCP
      name: "api"
      port: 80
      targetPort: 9376
```

## Description

The sidecar proxy needs to understand the protocols it is proxying so that it
can apply policies. Istio Pilot decides which protocol is in use by examining
the service port name and checking if it is named like `<protocol>-<anything
else>`. For example, `https-api` (treated as the `https` protocol) or
`http2-monitoring` (treated as the `http2` protocol).

## Suggested Resolution

Rename the target services target port(s) to 'http', 'http2', 'https', or prefix with 'http-', 'http2-', 'https-'. 

```yaml
apiVersion: v1
kind: Service
metadata:
  name: platform-api
  namespace: platform
spec:
  selector:
    app: platform-api
  ports:
    - protocol: TCP
      name: "https-api"
      port: 80
      targetPort: 9376
```

## See Also

- [Pod and Service Requirements](https://istio.io/docs/setup/kubernetes/prepare/requirements/)
- [End-user authentication](https://istio.io/docs/tasks/security/authn-policy/#end-user-authentication)

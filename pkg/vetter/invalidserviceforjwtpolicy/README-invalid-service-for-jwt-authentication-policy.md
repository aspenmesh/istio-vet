# Invalid Service For JWT Authentication Policy

## Example

The JWT Authentication Policy `your-jwt-authpolicy` in namespace `your-apps` targets service `your-service` in the same
namespace. The target services port names are not prefixed with mesh supported protocols.

## Description

## Suggested Resolution

Rename the target services target port(s) to 'http', 'http2', 'https', or prefix with 'http-', 'http2-', 'https-'. 

## See Also

- [Pod and Service Requirements](https://istio.io/docs/setup/kubernetes/prepare/requirements/)

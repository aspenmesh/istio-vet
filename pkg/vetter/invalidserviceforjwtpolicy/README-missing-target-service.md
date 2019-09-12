# Target Service Not Found For JWT Authentication Policy

## Example

The JWT Authentication Policy `platform-jwt-authpolicy` in namespace `platform` targets service `platform-api` in the same
namespace. The target service does not exist.

## Description

Authentication Policies can only target services in the same namespace.

## Suggested Resolution

Ensure the target service is in the same namespace as the Authentication Policy and that the name is correct.

## See Also

- [End-user authentication](https://istio.io/docs/tasks/security/authn-policy/#end-user-authentication)
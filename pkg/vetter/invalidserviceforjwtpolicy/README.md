# Invalid Service For JWT Authentication Policy

The `InvalidServiceForJWTPolicy` vetter inspects the port names defined in the Authentication Policies target Service and 
generates notes if they are missing the following Istio recognized port protocol prefixes:

* http
* http2
* https

## Notes Generated

- [Invalid Service For JWT Authentication Policy](README-invalid-service-for-jwt-authentication-policy.md)
- [Target Service Not Found For JWT Authentication Policy](README-target-service-not-found-for-jwt-authentication-policy.md)

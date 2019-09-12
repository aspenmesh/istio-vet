# Invalid Service For JWT Authentication Policy

The `InvalidServiceForJWTPolicy` vetter inspects the port names defined in the Authentication Policies target Service and 
generates notes if they are missing the following Istio recognized port protocol prefixes:

* http
* http2
* https

## Notes Generated

- [Invalid Service For JWT Authentication Policy](README-invalid-target-service-port-name.md)
- [Target Service Not Found For JWT Authentication Policy](README-missing-target-service.md)

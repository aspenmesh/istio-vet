# Invalid Service For JWT Authentication Policy
The `InvalidServiceForJWTPolicy` vetter inspects the Authentication Policies and looks for the following misconfigurations:

- Invalid target service port names - Generates notes if the target service is missing the following Istio recognized port protocol prefixes:
    - http
    - http2
    - https
- Missing target service - Generates a note if the target service can not be found in the same namespace as the Authentication Policy.

## Notes Generated

- [Invalid Service For JWT Authentication Policy](README-invalid-target-service-port-name.md)
- [Target Service Not Found For JWT Authentication Policy](README-missing-target-service.md)

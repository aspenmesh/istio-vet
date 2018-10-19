/*
Copyright 2018 Aspen Mesh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mtlspolicyutil

import (
	"errors"
	"strings"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	istioauthv1alpha1 "istio.io/api/authentication/v1alpha1"
)

// State of the mTLS settings
type MTLSSetting int32

const (
	// Unknown if state cannot be determined
	MTLSSetting_UNKNOWN MTLSSetting = 0
	// Enabled if mTLS is turned on
	MTLSSetting_ENABLED MTLSSetting = 1
	// Disabled if mTLS is turned off
	MTLSSetting_DISABLED MTLSSetting = 2
	// Mixed if mTLS is partially enabled or disabled
	MTLSSetting_MIXED MTLSSetting = 3
)

type policiesByNamespaceMap map[string][]*authv1alpha1.Policy
type policiesByNameMap map[string][]*authv1alpha1.Policy
type policiesByNamespaceNameMap map[string]policiesByNameMap
type policiesByPortMap map[uint32][]*authv1alpha1.Policy
type policiesByNamePortMap map[string]policiesByPortMap
type policiesByNamespaceNamePortMap map[string]policiesByNamePortMap

// AuthPolicies holds maps of Istio authorization policies by port, name, namespace
type AuthPolicies struct {
	mesh      []*authv1alpha1.MeshPolicy
	namespace policiesByNamespaceMap
	name      policiesByNamespaceNameMap
	port      policiesByNamespaceNamePortMap
}

// NewAuthPolicies initializes the maps for an AuthPolicies to be loaded by
// LoadAuthPolicies
func NewAuthPolicies() *AuthPolicies {
	return &AuthPolicies{
		mesh:      []*authv1alpha1.MeshPolicy{},
		namespace: make(policiesByNamespaceMap),
		name:      make(policiesByNamespaceNameMap),
		port:      make(policiesByNamespaceNamePortMap),
	}
}

// AddByMesh adds a Policy to the AuthPolicies mesh map
func (ap *AuthPolicies) AddByMesh(mp *authv1alpha1.MeshPolicy) {
	ap.mesh = append(ap.mesh, mp)
}

// AddByNamespace adds a Policy to the AuthPolicies namespace map
func (ap *AuthPolicies) AddByNamespace(namespace string, policy *authv1alpha1.Policy) {
	n := ap.namespace[namespace]
	ap.namespace[namespace] = append(n, policy)
}

// AddByName adds a Policy to the AuthPolicies name map
func (ap *AuthPolicies) AddByName(s Service, policy *authv1alpha1.Policy) {
	namespace, ok := ap.name[s.Namespace]
	if !ok {
		namespace = make(policiesByNameMap)
		ap.name[s.Namespace] = namespace
	}
	name, _ := namespace[s.Name]
	namespace[s.Name] = append(name, policy)
}

// AddByPort adds a Policy to the AuthPolicies port map
func (ap *AuthPolicies) AddByPort(s Service, port uint32, policy *authv1alpha1.Policy) {
	namespace, ok := ap.port[s.Namespace]
	if !ok {
		namespace = make(policiesByNamePortMap)
		ap.port[s.Namespace] = namespace
	}
	name, ok := namespace[s.Name]
	if !ok {
		name = make(policiesByPortMap)
		namespace[s.Name] = name
	}
	p, _ := name[port]
	name[port] = append(p, policy)
}

// ByMesh is passed a mesh and returns the Policy in the AuthPolicies
// mesh map for that mesh
func (ap *AuthPolicies) ByMesh() []*authv1alpha1.MeshPolicy {
	return ap.mesh
}

// ByNamespace is passed a namespace and returns the Policy in the AuthPolicies
// namespace map for that namespace
func (ap *AuthPolicies) ByNamespace(namespace string) []*authv1alpha1.Policy {
	n, ok := ap.namespace[namespace]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	return n
}

// ByName is passed a Service and returns the Policy in the AuthPolicies
// namespace map for the name of that Service
func (ap *AuthPolicies) ByName(s Service) []*authv1alpha1.Policy {
	ns, ok := ap.name[s.Namespace]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	n, ok := ns[s.Name]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	return n
}

// ByPort is passed a Service and a port number and returns the Policy in the
// AuthPolicies port map for that port number
func (ap *AuthPolicies) ByPort(s Service, port uint32) []*authv1alpha1.Policy {
	ns, ok := ap.port[s.Namespace]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	n, ok := ns[s.Name]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	p, ok := n[port]
	if !ok {
		return []*authv1alpha1.Policy{}
	}
	return p
}

// ForEachPolByPort takes a callback and applies it to a range of policies by port
func (ap *AuthPolicies) ForEachPolByPort(s Service, cb func(policies []*authv1alpha1.Policy)) {
	nsPortPols, ok := ap.port[s.Namespace]
	if !ok {
		return
	}
	nPortPols, ok := nsPortPols[s.Name]
	if !ok {
		return
	}
	for _, policies := range nPortPols {
		cb(policies)
	}
}

// getMTLSBool returns a bool and error from the 4 possible enum mTls states.
// Mixed counts as enabled since it allows enabled traffic, but it returns an error in case the caller needs to know if the true status means it's enabled-only, or enabled in a way that allows other traffic.
// Unknown counts as disabled since we cannot tell the caller that the status is mTls enabled. It returns an error in case the caller needs to know if the false status means that the false status is actually bogus because we we unable to determine the mTls status.
func getMTLSBool(mtlsState MTLSSetting) bool {
	// pass in the policy to maintain the structure of returns for callers pre-Oct2018-refactor.
	switch checkState := mtlsState; {
	case checkState == MTLSSetting_ENABLED:
		return true
	case checkState == MTLSSetting_UNKNOWN:
		return false
	case checkState == MTLSSetting_MIXED:
		return true
	default:
		return false
	}
}

// paramIsMTls determines whether mTls is enabled for a policy when no modes are listed.
// If a yaml file contains "- mtls: ", the Policy Object will be `"spec":{"peers":[{"mtls":null}]}` Istio Docs describe this as mTls-DISABLED.
// If a yaml file contains "- mtls: {}", the Policy Object will be `"spec":{"peers":[{"mtls": {}}]}` Istio Docs describe this as mTls-ENABLED.
// We can't use .GetMtls() because it will return nil in cases where the peer isn't mTls as well as in cases where mtls is listed but empty. paramIsMTls() checks whether the peer list's params can be coerced into a PeerAuthenticationMethod_Mtls which accounts for a null versus {} mtls entry.
func paramIsMTls(peer *istioauthv1alpha1.PeerAuthenticationMethod) bool {
	_, ok := peer.GetParams().(*istioauthv1alpha1.PeerAuthenticationMethod_Mtls)
	if ok {
		// `"spec":{"peers":[{"mtls": {}}]}` means enabled
		return true
	}
	// `"spec":{"peers":[{"mtls":null}]}` means disabled
	return false
}

// getModeFromPeers takes a set of peers and returns an mTls setting
// Once passed a set of peerAuthMethods, getModeFromPeers() checks whether each is an mTls setting or a JWT, then if it's mTls, it checks the mode for its setting.
func getModeFromPeers(peerAuthMethods []*istioauthv1alpha1.PeerAuthenticationMethod) MTLSSetting {
	var mtlsState MTLSSetting
	// peerAuthMethods is checked for being Empty in the calling function.

	// Per peerAuthMethod, check if it lists mtls in any way, then check for mtls Mode. Count the number of enabled or mixed methods to determine the final mtls state for this policy.
	var enabled, mixed int
	for _, pam := range peerAuthMethods {
		// PeerAuthenticationMethod could be JWT or multiple mtls settings.
		if pam.GetMtls() != nil {
			// A peer sections exists with mTls enabled, so check its Mode for STRICT or PERMISSIVE.
			peerMode := pam.GetMtls().GetMode()
			if peerMode == istioauthv1alpha1.MutualTls_STRICT {
				enabled++
			} else {
				mixed++
			}
		} else {
			// A peer section exists, but the peer authentication methods are the odd cases where "- mtls: {}" or "- mtls : ", so GetMtls() will return nil even though Istio considers "- mtls: {}" to be enabled.  paramIsMTls() checks for these both cases.
			if paramIsMTls(pam) {
				enabled++
			}
		}
	}
	// If there is any occurrance of mixed, nothing else matters.
	// If !mixed, check for enabled
	// DISABLED will be returned in cases where there is a Peer Authentication Method, but it is JWT instead of mTls and there is no other mTls setting for the policy
	if mixed != 0 {
		mtlsState = MTLSSetting_MIXED
	} else if enabled != 0 {
		mtlsState = MTLSSetting_ENABLED
	} else {
		mtlsState = MTLSSetting_DISABLED
	}
	return mtlsState
}

// evaluateMTlsForPeer takes a set of peers and the peerIsOptional setting, and returns an mTls setting.
// Once passed a set of peers and peerIsOptional setting, it returns the determined mTls state, or calls getModeFromPeers() to check its mTls state.
func evaluateMTlsForPeer(peers []*istioauthv1alpha1.PeerAuthenticationMethod, peerOptional bool) MTLSSetting {
	var mtlsState MTLSSetting
	// Check to see if Peers has a list of peerAuthMethods or is empty.
	if len(peers) == 0 {
		// If Peers exists && is empty, Istio considers it mtls-disabled
		mtlsState = MTLSSetting_DISABLED
	} else if peerOptional == true {
		// If Peers has at least one item in the list, check to see if the user set PeerIsOptional == true. If so, this overrides any other mtls settings. Functionality is broken in Istio1.0, but is fixed as of Istio 1.3
		mtlsState = MTLSSetting_MIXED
	} else {
		mtlsState = getModeFromPeers(peers)
	}
	return mtlsState
}

// MeshPolicyIsMtls returns true if the passed Policy has mTLS enabled.
// The duplicaiton in code for MeshPolicyIsMtls and AuthPolicyIsMtls is because the two objects are a different type and cannot use the same code. Once peers have been accessed, the two kinds of policy can use the same code.
func MeshPolicyIsMtls(policy *authv1alpha1.MeshPolicy) MTLSSetting {
	peers := policy.Spec.GetPeers()
	if peers == nil {
		return MTLSSetting_DISABLED
	}
	peerOptional := policy.Spec.GetPeerIsOptional()
	return evaluateMTlsForPeer(peers, peerOptional)
}

// AuthPolicyIsMtls returns true if the passed Policy has mTLS enabled.
// The duplicaiton in code for MeshPolicyIsMtls and AuthPolicyIsMtls is because the two objects are a different type and cannot use the same code. Once peers have been accessed, the two kinds of policy can use the same code.
func AuthPolicyIsMtls(policy *authv1alpha1.Policy) MTLSSetting {
	peers := policy.Spec.GetPeers()
	if peers == nil {
		return MTLSSetting_DISABLED
	}
	peerOptional := policy.Spec.GetPeerIsOptional()
	return evaluateMTlsForPeer(peers, peerOptional)
}

// IsGlobalMtlsEnabled validates that there are the expected number of
// MeshPolicies in the list (0 or 1), validates the name of the MeshPolicy, and
// returns true if the MeshPolicy enables mTLS
func IsGlobalMtlsEnabled(meshPolicies []*authv1alpha1.MeshPolicy) (bool, error) {
	if len(meshPolicies) > 1 {
		return false, errors.New("More than one MeshPolicy was found")
	} else if len(meshPolicies) == 0 {
		return false, nil
	} else {
		if strings.EqualFold(meshPolicies[0].ObjectMeta.Name, "default") {
			mtlsState := MeshPolicyIsMtls(meshPolicies[0])
			ok := getMTLSBool(mtlsState)
			return ok, nil
		} else {
			return false, errors.New("MeshPolicy is not named 'default'")
		}
	}
}

// TLSDetailsByPort walks through Auth Policies at the port level and returns the mtlsState for the requested resource. It returns the mTls state for the parent resource if there is no policy for the requested resource.
func (ap *AuthPolicies) TLSDetailsByPort(s Service, port uint32) (MTLSSetting, *authv1alpha1.Policy, error) {
	policies := ap.ByPort(s, port)
	if len(policies) > 1 {
		// TODO (BLaurenB) We think that in Istio 0.8, non-conflicting policies (or identical policies) will "work" because the behavior requested will be the same regardless of which one Istio chooses. We may need to handle this differently.
		return MTLSSetting_UNKNOWN, nil, errors.New("Conflicting policies for port")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// If there are no policies for the port, return mtlsState for parent resource.
	return ap.TLSDetailsByName(s)
}

// TLSByPort wraps TLSDetailsByPort and returns a boolean.
func (ap *AuthPolicies) TLSByPort(s Service, port uint32) (bool, *authv1alpha1.Policy, error) {
	mtlsState, policy, err := ap.TLSDetailsByPort(s, port)
	if err != nil {
		// The false status is actually bogus because we we unable to determine the mTls status.
		return false, nil, err
	}
	return getMTLSBool(mtlsState), policy, err
}

// TLSDetailsByName walks through Auth Policies at the port and name level, and returns the mtlsState for the requested resource. It returns the mTls state for the parent resource if there is no policy for the requested resource.
func (ap *AuthPolicies) TLSDetailsByName(s Service) (MTLSSetting, *authv1alpha1.Policy, error) {

	policies := ap.ByName(s)
	if len(policies) > 1 {
		// TODO (BLaurenB) We think that in Istio 0.8, non-conflicting policies (or identical policies) will "work" because the behavior requested will be the same regardless of which one Istio chooses. We may need to handle this differently.
		return MTLSSetting_UNKNOWN, nil, errors.New("Conflicting policies for service by name")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// If there are no policies for the service, return mtlsState for parent resource.
	return ap.TLSDetailsByNamespace(s)
}

// TLSByName wraps TLSDetailsByName and returns a boolean.
func (ap *AuthPolicies) TLSByName(s Service) (bool, *authv1alpha1.Policy, error) {
	mtlsState, policy, err := ap.TLSDetailsByName(s)
	if err != nil {
		// The false status is actually bogus because we we unable to determine the mTls status.
		return false, nil, err
	}
	return getMTLSBool(mtlsState), policy, err
}

// TLSDetailsByNamespace walks through Auth Policies at the port, name, and namespace level and returns the mtlsState for the requested resource. It returns the mTls state for the parent resource if there is no policy for the requested resource.
func (ap *AuthPolicies) TLSDetailsByNamespace(s Service) (MTLSSetting, *authv1alpha1.Policy, error) {
	policies := ap.ByNamespace(s.Namespace)
	if len(policies) > 1 {
		// TODO (BLaurenB) We think that in Istio 0.8, non-conflicting policies (or identical policies) will "work" because the behavior requested will be the same regardless of which one Istio chooses. We may need to handle this differently.
		return MTLSSetting_UNKNOWN, nil, errors.New("Conflicting policies for service by namespace")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// If there are no policies for the namespace, return mtlsState for parent resource. Note this function can't return a Mesh Policy since it's a different Type.
	mtlsState, _, err := ap.TLSDetailsByMesh()
	return mtlsState, nil, err
}

// TLSByNamespace wraps TLSDetailsByNamespace and returns a boolean.
func (ap *AuthPolicies) TLSByNamespace(s Service) (bool, *authv1alpha1.Policy, error) {
	mtlsState, policy, err := ap.TLSDetailsByNamespace(s)
	if err != nil {
		// The false status is actually bogus because we we unable to determine the mTls status.
		return false, nil, err
	}
	return getMTLSBool(mtlsState), policy, err
}

func (ap *AuthPolicies) TLSDetailsByMesh() (MTLSSetting, *authv1alpha1.MeshPolicy, error) {
	policies := ap.ByMesh()
	if len(policies) > 1 {
		// There can be only one Mesh policy and it must be named "default"
		return MTLSSetting_UNKNOWN, nil, errors.New("Conflicting policies for service by mesh")
	}
	if len(policies) == 1 {
		return MeshPolicyIsMtls(policies[0]), policies[0], nil
	}
	// If there is no mesh policy, mTls is considered to be disabled for the cluster.
	return MTLSSetting_DISABLED, nil, nil
}

// LoadAuthPolicies is passed a list of Policies and returns an
// AuthPolicies struct with maps of policies by port, name, and namespace.
// The function separates the policies so that the namespace map only includes policies that are namespace-wide only, service map includes policies that are service-wide only, and port map includes policies that designate a target port.
func LoadAuthPolicies(policies []*authv1alpha1.Policy,
	meshPolicies []*authv1alpha1.MeshPolicy) (*AuthPolicies, error) {
	loaded := NewAuthPolicies()
	for _, mp := range meshPolicies {
		loaded.AddByMesh(mp)
	}
	for _, policy := range policies {
		targets := policy.Spec.GetTargets()
		if targets == nil || len(targets) == 0 {
			// No targets: this is a namespace-wide policy.
			if policy.Name != "default" {
				// This policy is invalid according to docs.
				continue
			}
			loaded.AddByNamespace(policy.Namespace, policy)
			continue
		}
		// Policy has targets.
		for _, target := range targets {
			name := target.GetName()
			if name == "" {
				// According to docs, this is invalid.
				continue
			}
			s := Service{Name: name, Namespace: policy.Namespace}
			ports := target.GetPorts()
			if ports == nil || len(ports) == 0 {
				// This policy applies to a service by name
				loaded.AddByName(s, policy)
				continue
			}

			for _, port := range ports {
				n := port.GetNumber()
				if n == 0 {
					continue
				}
				loaded.AddByPort(s, n, policy)
			}
		}
	}
	return loaded, nil
}

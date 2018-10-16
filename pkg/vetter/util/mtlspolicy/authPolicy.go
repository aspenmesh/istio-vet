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

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	istioauthv1alpha1 "istio.io/api/authentication/v1alpha1"
)

// State of the mTLS settings
type MutualTLSSetting int32

const (
	meshName = "mesh"
	// Unknown if state cannot be determined
	MutualTLSSetting_UNKNOWN MutualTLSSetting = 0
	// Enabled if mTLS is turned on
	MutualTLSSetting_ENABLED MutualTLSSetting = 1
	// Disabled if mTLS is turned off
	MutualTLSSetting_DISABLED MutualTLSSetting = 2
	// Mixed if mTLS is partially enabled or disabled
	MutualTLSSetting_MIXED MutualTLSSetting = 3
)

type policiesByMeshMap map[string][]*authv1alpha1.MeshPolicy
type policiesByNamespaceMap map[string][]*authv1alpha1.Policy
type policiesByNameMap map[string][]*authv1alpha1.Policy
type policiesByNamespaceNameMap map[string]policiesByNameMap
type policiesByPortMap map[uint32][]*authv1alpha1.Policy
type policiesByNamePortMap map[string]policiesByPortMap
type policiesByNamespaceNamePortMap map[string]policiesByNamePortMap

// AuthPolicies holds maps of Istio authorization policies by port, name, namespace
type AuthPolicies struct {
	mesh      policiesByMeshMap
	namespace policiesByNamespaceMap
	name      policiesByNamespaceNameMap
	port      policiesByNamespaceNamePortMap
}

// NewAuthPolicies initializes the maps for an AuthPolicies to be loaded by
// LoadAuthPolicies
func NewAuthPolicies() *AuthPolicies {
	return &AuthPolicies{
		mesh:      make(policiesByMeshMap),
		namespace: make(policiesByNamespaceMap),
		name:      make(policiesByNamespaceNameMap),
		port:      make(policiesByNamespaceNamePortMap),
	}
}

// AddByMesh adds a Policy to the AuthPolicies mesh map
func (ap *AuthPolicies) AddByMesh(mesh string, policy *authv1alpha1.MeshPolicy) {
	m := ap.mesh[mesh]
	ap.mesh[mesh] = append(m, policy)
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
	// Currently only UI for 1 cluster. MeshPolicy must be named "default".
	m, ok := ap.mesh[meshName]
	if !ok {
		return []*authv1alpha1.MeshPolicy{}
	}
	return m
}

// ByNamespace is passed a namespace and returns the Policy in the AuthPolicies
// namespace map for that namespace
// If passed an empty string, it will return the whole map of policies by namespace
func (ap *AuthPolicies) ByNamespace(namespace string) ([]*authv1alpha1.Policy, policiesByNamespaceMap) {
	if namespace == "" {
		return nil, ap.namespace
	}
	n, ok := ap.namespace[namespace]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	return n, nil
}

// ByName is passed a Service and returns the Policy in the AuthPolicies
// namespace map for the name of that Service
// If passed an empty string for s.Name and s.Namespace, it will return the map of policies by namespace[service].
func (ap *AuthPolicies) ByName(s Service) ([]*authv1alpha1.Policy, policiesByNamespaceNameMap) {
	if s.Name == "" && s.Namespace == "" {
		return nil, ap.name
	}
	ns, ok := ap.name[s.Namespace]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	n, ok := ns[s.Name]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	return n, nil
}

// ByPort is passed a Service and a port number and returns the Policy in the
// AuthPolicies port map for that port number
// If passed zero as the port number, it will return the map of all policies with a port as a target.
func (ap *AuthPolicies) ByPort(s Service, port uint32) ([]*authv1alpha1.Policy, policiesByNamespaceNamePortMap) {
	if port == 0 {
		return nil, ap.port
	}
	ns, ok := ap.port[s.Namespace]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	n, ok := ns[s.Name]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	p, ok := n[port]
	if !ok {
		return []*authv1alpha1.Policy{}, nil
	}
	return p, nil
}

// paramIsMTls determines whether mTls is enabled for a policy when no modes are listed.
// If a yaml file contains "- mtls: {}" or "- mtls: ", the Policy Object will be `spec":{"peers":[{"mtls":null}]}}` Istio Docs describe this as mTls-enabled. We can't use .GetMtls() because it will return nil in cases where the peer isn't mTls as well as in cases where mtls is listed but empty
func paramIsMTls(peer *istioauthv1alpha1.PeerAuthenticationMethod) bool {
	_, ok := peer.GetParams().(*istioauthv1alpha1.PeerAuthenticationMethod_Mtls)
	if ok {
		return true
	}
	return false
}

// getModeFromPeers takes a set of peers and returns an mTls setting
// Once passed a set of peerAuthMethods, getModeFromPeers() checks whether each is an mTls setting or a JWT, then if it's mTls, it checks the mode for its setting.
func getModeFromPeers(peerAuthMethods []*istioauthv1alpha1.PeerAuthenticationMethod) MutualTLSSetting {
	var mtlsState MutualTLSSetting
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
			// A peer section exists, but the peer authentication methods are the odd cases where "- mtls: {}" or "- mtls : ", so GetMtls() will return nil even though Istio considers these cases to be enabled.  paramIsMTls() checks for these enabled cases.
			if paramIsMTls(pam) {
				enabled++
			}
		}
	}
	// If there is any occurrance of mixed, nothing else matters.
	// If !mixed, check for enabled
	// DISABLED will be returned in cases where there is a Peer Authentication Method, but it is JWT instead of mTls and there is no other mTls setting for the policy
	if mixed != 0 {
		mtlsState = MutualTLSSetting_MIXED
	} else if enabled != 0 {
		mtlsState = MutualTLSSetting_ENABLED
	} else {
		mtlsState = MutualTLSSetting_DISABLED
	}
	return mtlsState
}

// evaluateMTlsForPeer takes a set of peets and the peerIsOptional setting, and returns an mTls setting.
// Once passed a set of peers and peerIsOptional setting, it returns the determined mTls state, or calls getModeFromPeers() to check its mTls state.
func evaluateMTlsForPeer(peers []*istioauthv1alpha1.PeerAuthenticationMethod, peerOptional bool) MutualTLSSetting {
	var mtlsState MutualTLSSetting
	// Check to see if Peers has a list of peerAuthMethods or is empty.
	if len(peers) == 0 {
		// If Peers exists && is empty, Istio considers it mtls-disabled
		mtlsState = MutualTLSSetting_DISABLED
	} else if peerOptional == true {
		// If Peers has at least one item in the list, check to see if the user set PeerIsOptional == true. If so, this overrides any other mtls settings. Functionality is broken in Istio1.0, but is fixed as of Istio 1.3
		mtlsState = MutualTLSSetting_MIXED
	} else {
		mtlsState = getModeFromPeers(peers)
	}
	return mtlsState
}

// MeshPolicyIsMtls returns true if the passed Policy has mTLS enabled.
// The duplicaiton in code for MeshPolicyIsMtls and AuthPolicyIsMtls is because the two objects are different and cannot use the same code. Once peers have been accessed, the two kinds of policy can use the same code.
func MeshPolicyIsMtls(policy *authv1alpha1.MeshPolicy) MutualTLSSetting {
	peers := policy.Spec.GetPeers()
	if peers == nil {
		return MutualTLSSetting_DISABLED
	}
	peerOptional := policy.Spec.GetPeerIsOptional()
	return evaluateMTlsForPeer(peers, peerOptional)
}

// AuthPolicyIsMtls returns true if the passed Policy has mTLS enabled.
// The duplicaiton in code for MeshPolicyIsMtls and AuthPolicyIsMtls is because the two objects are different and cannot use the same code. Once peers have been accessed, the two kinds of policy can use the same code.
func AuthPolicyIsMtls(policy *authv1alpha1.Policy) MutualTLSSetting {
	peers := policy.Spec.GetPeers()
	if peers == nil {
		return MutualTLSSetting_DISABLED
	}
	peerOptional := policy.Spec.GetPeerIsOptional()
	return evaluateMTlsForPeer(peers, peerOptional)
}

// TLSDetailsByPort walks through Auth Policies at the port, name, and namespace level and returns the mtlsState for the requested resource
func (ap *AuthPolicies) TLSDetailsByPort(s Service, port uint32) (MutualTLSSetting, *authv1alpha1.Policy, error) {
	policies, _ := ap.ByPort(s, port)
	if len(policies) > 1 {
		// TODO(m-eaton): If all the policies are the same, does it work?
		return MutualTLSSetting_UNKNOWN, nil, errors.New("Conflicting policies for port")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// if there are no policies for the port, return mtlsState for parent resource.
	return ap.TLSDetailsByName(s)
}

// TLSByPort wraps TLSDetailsByPort and returns a boolean.
func (ap *AuthPolicies) TLSByPort(s Service, port uint32) (bool, error) {
	mtlsState, _, err := ap.TLSDetailsByPort(s, port)
	if mtlsState == MutualTLSSetting_ENABLED {
		return true, nil
	} else if mtlsState == MutualTLSSetting_UNKNOWN {
		return false, errors.New("mTLS status is unknown")
	}
	return false, nil
}

// TLSByName walks through Policies at the name and namespace level and
// returns true if the Policy found has mTLS enabled
func (ap *AuthPolicies) TLSByName(s Service) (bool, *authv1alpha1.Policy, error) {
	policies, _ := ap.ByName(s)
	if len(policies) > 1 {
		// TODO: If all the policies are the same, does it work?
		return false, nil, errors.New("Conflicting policies for service by name")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	return ap.TLSByNamespace(s)
}

// TLSByNamespace walks through Policies at the namespace level and
// returns true if the Policy found has mTLS enabled
func (ap *AuthPolicies) TLSByNamespace(s Service) (bool, *authv1alpha1.Policy, error) {
	policies, _ := ap.ByNamespace(s.Namespace)
	if len(policies) > 1 {
		// TODO: If all the policies are the same, does it work?
		return false, nil, errors.New("Conflicting policies for service by namespace")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	return false, nil, errors.New("No policy for service by namespace")
}

func (ap *AuthPolicies) TLSByMesh() (bool, *authv1alpha1.MeshPolicy, error) {
	policies := ap.ByMesh()
	if len(policies) > 1 {
		// There can be only one Mesh policy and it must be named "default"
		return false, nil, errors.New("Conflicting policies for service by mesh")
	}
	if len(policies) == 1 {
		return MeshPolicyIsMtls(policies[0]), policies[0], nil
	}
	return false, nil, errors.New("No policy for service by mesh")
}

func (ap *AuthPolicies) LoadMeshPolicy(policies []*authv1alpha1.MeshPolicy) *AuthPolicies {
	for _, policy := range policies {
		if policy.Name != "default" {
			// Mesh Policy must be named "default".
			continue
		}
		ap.AddByMesh(meshName, policy)
		continue
	}
	return ap
}

// LoadAuthPolicies is passed a list of Policies and returns an AuthPolicies
// with each of the Policies mapped by port, name, and namespace
func LoadAuthPolicies(policies []*authv1alpha1.Policy) (*AuthPolicies, error) {
	loaded := NewAuthPolicies()
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

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

type policiesByNamespaceMap map[string][]*authv1alpha1.Policy
type policiesByNameMap map[string][]*authv1alpha1.Policy
type policiesByNamespaceNameMap map[string]policiesByNameMap
type policiesByPortMap map[uint32][]*authv1alpha1.Policy
type policiesByNamePortMap map[string]policiesByPortMap
type policiesByNamespaceNamePortMap map[string]policiesByNamePortMap

// AuthPolicies holds maps of Istio authorization policies by port, name, namespace
type AuthPolicies struct {
	namespace policiesByNamespaceMap
	name      policiesByNamespaceNameMap
	port      policiesByNamespaceNamePortMap
}

// NewAuthPolicies initializes the maps for an AuthPolicies to be loaded by
// LoadAuthPolicies
func NewAuthPolicies() *AuthPolicies {
	return &AuthPolicies{
		namespace: make(policiesByNamespaceMap),
		name:      make(policiesByNamespaceNameMap),
		port:      make(policiesByNamespaceNamePortMap),
	}
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

// AuthPolicyIsMtls returns true if the passed Policy has mTLS enabled
func AuthPolicyIsMtls(policy *authv1alpha1.Policy) bool {
	peers := policy.Spec.GetPeers()
	if peers == nil {
		return false
	}
	for _, peer := range peers {
		// mTLS is "on" if there is an mTLS peer entry, even if it is nil.
		// so e.g.:
		//   peers:
		//   - mtls: null
		// We can't use .GetMtls(), we need to attempt the cast ourselves, because
		// .GetMtls() will return nil if the peer isn't mTLS AND if the peer is an
		// empty mTLS, and we won't be able to distinguish.
		_, ok := peer.GetParams().(*istioauthv1alpha1.PeerAuthenticationMethod_Mtls)
		if ok {
			return true
		}
	}
	return false
}

// TLSByPort walks through Policies at the port level and
// returns true if the Policy found has mTLS enabled
func (ap *AuthPolicies) TLSByPort(s Service, port uint32) (bool, *authv1alpha1.Policy, error) {
	policies := ap.ByPort(s, port)
	if len(policies) > 1 {
		// TODO: If all the policies are the same, does it work?
		return false, nil, errors.New("Conflicting policies for port")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// TODO: Walk the next tier?
	return false, nil, errors.New("No policy for port")
}

// TLSByName walks through Policies at the name level and
// returns true if the Policy found has mTLS enabled
func (ap *AuthPolicies) TLSByName(s Service) (bool, *authv1alpha1.Policy, error) {
	policies := ap.ByName(s)
	if len(policies) > 1 {
		// TODO: If all the policies are the same, does it work?
		return false, nil, errors.New("Conflicting policies for service by name")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	// TODO: Walk the next tier?
	return false, nil, errors.New("No policy for service by name")
}

// TLSByNamespace walks through Policies at the namespace level and
// returns true if the Policy found has mTLS enabled
func (ap *AuthPolicies) TLSByNamespace(s Service) (bool, *authv1alpha1.Policy, error) {
	policies := ap.ByNamespace(s.Namespace)
	if len(policies) > 1 {
		// TODO: If all the policies are the same, does it work?
		return false, nil, errors.New("Conflicting policies for service by namespace")
	}
	if len(policies) == 1 {
		return AuthPolicyIsMtls(policies[0]), policies[0], nil
	}
	return false, nil, errors.New("No policy for service by namespace")
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

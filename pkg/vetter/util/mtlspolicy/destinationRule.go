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

	istioNet "istio.io/api/networking/v1beta1"
	istioClientNet "istio.io/client-go/pkg/apis/networking/v1beta1"
)

// Destination Rules can have arbitrary PortTrafficPolicy; we don't want to
// re-walk a destination rule to find the correct PortTrafficPolicy for the
// particular port we're working on.  For ports, store both the
// DestinationRule, and the PortTrafficPolicy inside that we care about for
// this particular port.

type destRulesByNamespaceMap map[string][]*istioClientNet.DestinationRule
type destRulesByNameMap map[string][]*istioClientNet.DestinationRule
type destRulesByNamespaceNameMap map[string]destRulesByNameMap
type destRulesByPortMap map[uint32][]*PortDestRule
type destRulesByNamePortMap map[string]destRulesByPortMap
type destRulesByNamespaceNamePortMap map[string]destRulesByNamePortMap

// DestRules holds maps of Istio destination rules by port, name, and namespace
type DestRules struct {
	// Destination rules must have a host specifier.  In Istio 1.0 they cannot be
	// cluster-wide, but they can be namespace wide
	// (*.namespace.svc.cluster.local)
	namespace destRulesByNamespaceMap
	name      destRulesByNamespaceNameMap
	port      destRulesByNamespaceNamePortMap
}

// PortDestRule stores the Istio destination rule and port traffic policy for a port
type PortDestRule struct {
	Rule     *istioClientNet.DestinationRule
	PortRule *istioNet.TrafficPolicy_PortTrafficPolicy
}

// NewDestRules initializes the maps for a DestRules to be loaded by
// LoadDestRules
func NewDestRules() *DestRules {
	return &DestRules{
		namespace: make(destRulesByNamespaceMap),
		name:      make(destRulesByNamespaceNameMap),
		port:      make(destRulesByNamespaceNamePortMap),
	}
}

// AddByNamespace adds a Destination Rule to the DestRules namespace map
func (dr *DestRules) AddByNamespace(namespace string, rule *istioClientNet.DestinationRule) {
	ns, _ := dr.namespace[namespace]
	dr.namespace[namespace] = append(ns, rule)
}

// AddByName adds a Destination Rule to the DestRules name map
func (dr *DestRules) AddByName(s Service, rule *istioClientNet.DestinationRule) {
	namespace, ok := dr.name[s.Namespace]
	if !ok {
		namespace = make(destRulesByNameMap)
		dr.name[s.Namespace] = namespace
	}
	name, _ := namespace[s.Name]
	namespace[s.Name] = append(name, rule)
}

// AddByPort adds a Destination Rule to the DestRules port map
func (dr *DestRules) AddByPort(
	s Service,
	port uint32,
	rule *istioClientNet.DestinationRule,
	portRule *istioNet.TrafficPolicy_PortTrafficPolicy,
) {
	namespace, ok := dr.port[s.Namespace]
	if !ok {
		namespace = make(destRulesByNamePortMap)
		dr.port[s.Namespace] = namespace
	}
	name, ok := namespace[s.Name]
	if !ok {
		name = make(destRulesByPortMap)
		namespace[s.Name] = name
	}
	p, _ := name[port]
	name[port] = append(p, &PortDestRule{Rule: rule, PortRule: portRule})
}

// ByNamespace is passed a namespace and returns the Destination Rule in the
// DestRules namespace map for that namespace
func (dr *DestRules) ByNamespace(namespace string) []*istioClientNet.DestinationRule {
	ns, ok := dr.namespace[namespace]
	if !ok {
		return []*istioClientNet.DestinationRule{}
	}
	return ns
}

// ByName is passed a Service and returns the Destination Rule in the
// DestRules name map for the name of that Service
func (dr *DestRules) ByName(s Service) []*istioClientNet.DestinationRule {
	ns, ok := dr.name[s.Namespace]
	if !ok {
		return []*istioClientNet.DestinationRule{}
	}
	res, ok := ns[s.Name]
	if !ok {
		return []*istioClientNet.DestinationRule{}
	}
	return res
}

// ByPort is passed a Service and a port number and returns the Destination Rule
// in the DestRules port map for that port number
func (dr *DestRules) ByPort(s Service, port uint32) []*PortDestRule {
	ns, ok := dr.port[s.Namespace]
	if !ok {
		return []*PortDestRule{}
	}
	n, ok := ns[s.Name]
	if !ok {
		return []*PortDestRule{}
	}
	res, ok := n[port]
	if !ok {
		return []*PortDestRule{}
	}
	return res
}

// ForEachByPort examines all Destination Rules for a Service and port number
// based off of a PortDestRule that is passed
func (dr *DestRules) ForEachByPort(cb func(s Service, port uint32, rule *PortDestRule)) {
	for namespace, rulesForNamespace := range dr.port {
		for name, rulesForName := range rulesForNamespace {
			s := Service{Name: name, Namespace: namespace}
			for port, rulesForPort := range rulesForName {
				for _, rule := range rulesForPort {
					cb(s, port, rule)
				}
			}
		}
	}
}

// ForEachByName examines all Destination Rules for a Service based off of a
// Destination Rule that is passed
func (dr *DestRules) ForEachByName(cb func(s Service, rule *istioClientNet.DestinationRule)) {
	for namespace, rulesForNamespace := range dr.name {
		for name, rulesForName := range rulesForNamespace {
			s := Service{Name: name, Namespace: namespace}
			for _, rule := range rulesForName {
				cb(s, rule)
			}
		}
	}
}

// PortDestRuleIsMtls returns true if mTLS is enabled for the PortDestRule
func PortDestRuleIsMtls(rule *PortDestRule) bool {
	return rule.PortRule.GetTls().GetMode() == istioNet.ClientTLSSettings_MUTUAL
}

// DestRuleIsMtls returns true if mTLS is enabled for the Destination Rule
func DestRuleIsMtls(rule *istioClientNet.DestinationRule) bool {
	return rule.Spec.GetTrafficPolicy().GetTls().GetMode() == istioNet.ClientTLSSettings_MUTUAL
}

// TLSByPort returns true if mTLS is enabled for the PortDestination rule of the
// port number passed
func (dr *DestRules) TLSByPort(s Service, port uint32) (bool, *PortDestRule, error) {
	rules := dr.ByPort(s, port)
	if len(rules) > 1 {
		// TODO: If all the rules are the same, does it work?
		return false, nil, errors.New("Conflicting rules for port")
	}
	if len(rules) == 1 {
		return PortDestRuleIsMtls(rules[0]), rules[0], nil
	}
	// TODO: Walk the next tier?
	return false, nil, errors.New("No rule for port")
}

// TLSByName returns true if mTLS is enabled for the Destination Rule
func (dr *DestRules) TLSByName(s Service) (bool, *istioClientNet.DestinationRule, error) {
	rules := dr.ByName(s)
	if len(rules) > 1 {
		// TODO: If all the rules are the same, does it work?
		return false, nil, errors.New("Conflicting rules for name")
	}
	if len(rules) == 1 {
		return DestRuleIsMtls(rules[0]), rules[0], nil
	}
	// TODO: Walk the next tier?
	return false, nil, errors.New("No rule for name")
}

// LoadDestRules is passed a list of Destination Rules and returns a DestRules
// with each of the Destination Rules mapped by port, name, and namespace
func LoadDestRules(rules []*istioClientNet.DestinationRule) (*DestRules, error) {
	loaded := NewDestRules()
	for _, r := range rules {
		host := r.Spec.Host
		if host == "" {
			// Host is REQUIRED according to Istio so skip this invalid rule
			continue
		}
		s, err := ServiceFromFqdn(host)
		if err != nil || r.Spec.GetTrafficPolicy() == nil {
			// Rule refers to a non-mesh service or has no TLS settings, skip.
			continue
		}
		// Handle the top-level policy
		if r.Spec.GetTrafficPolicy().GetTls() != nil {
			if s.Name == "*" {
				if s.Namespace == "*" {
					// This isn't allowed in Istio 1.0
					// TODO(andrew): Check that it doesn't work in 1.0.
					continue
				}
				loaded.AddByNamespace(s.Namespace, r)
			} else {
				loaded.AddByName(s, r)
			}
		}

		// For each port-level setting, handle the port specific overrides.
		for _, pl := range r.Spec.GetTrafficPolicy().GetPortLevelSettings() {
			if pl.Tls == nil || pl.Port == nil {
				// Rule has no TLS settings or no selector, skip
				continue
			}
			n := pl.GetPort().GetNumber()
			if n == 0 {
				continue
			}
			loaded.AddByPort(s, n, r, pl)
		}
	}
	return loaded, nil
}

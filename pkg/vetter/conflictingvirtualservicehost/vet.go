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

package conflictingvirtualservicehost

import (
	"fmt"
	"regexp"
	"strings"

	v1alpha3 "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "github.com/aspenmesh/istio-client-go/pkg/client/listers/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	v1 "k8s.io/client-go/listers/core/v1"
)

const (
	defaultGateway = "mesh"
	vetterID       = "ConflictingVirtualServiceHost"
	vsHostNoteType = "host-in-multiple-vs"
	vsHostSummary  = "Multiple VirtualServices define the same host (${host}) and gateway (${gateway}) and conflict"
	vsHostMsg      = "The VirtualServices ${vs_names} with routes ${routes}" +
		" define the same host (${host}) and gateway (${gateway}) and conflict. A VirtualService must have a unique combination of host and gateway or must not conflict." +
		" Consider updating the VirtualServices to have unique hostname and gateway or remove one of the conflicting rules."
)

type routeRuleType int

const (
	prefix routeRuleType = iota
	exact
	regex
)

// We need a type that is a Note with the keys
// that occur in this note "unfolded" (since
// we want to use that type as a key in a map,
// but the Attr field in Note is a map and hence
// unhashable)
type conflictingVsNote struct {
	Type    string
	Summary string
	Msg     string
	Level   apiv1.NoteLevel
	vsNames string
	gateway string
	routes  string
}

// VsHost implements Vetter interface
type VsHost struct {
	nsLister v1.NamespaceLister
	vsLister netv1alpha3.VirtualServiceLister
}
type hostAndGateway struct {
	gateway  string
	hostname string
}

type routeRule struct {
	ruleType  routeRuleType
	route     string
	vsName    string
	namespace string
}

type routeTrie struct {
	subRoutes  map[string]*routeTrie
	regexs     []routeRule
	routeRules []routeRule
}

type VirtualSvcByHostAndGateway map[hostAndGateway][]*v1alpha3.VirtualService

func asString(rrType routeRuleType) string {
	if rrType == prefix {
		return "prefix"
	} else if rrType == exact {
		return "exact"
	} else {
		return "regex"
	}
}

func unwrapNote(note conflictingVsNote, hosts []string) *apiv1.Note {
	return &apiv1.Note{
		Type:    note.Type,
		Summary: note.Summary,
		Msg:     note.Msg,
		Level:   apiv1.NoteLevel_ERROR,
		Attr: map[string]string{
			"vs_names": note.vsNames,
			"host":     strings.Join(hosts, " "),
			"gateway":  note.gateway,
			"routes":   note.routes,
		}}

}

// CreateVirtualServiceNotes checks for multiple vs defining the same host and
// generates notes for these cases
func CreateVirtualServiceNotes(virtualServices []*v1alpha3.VirtualService) ([]*apiv1.Note, error) {
	vsByHostAndGateway := VirtualSvcByHostAndGateway{}
	for _, vs := range virtualServices {
		for _, host := range vs.Spec.GetHosts() {
			h, err := util.ConvertHostnameToFQDN(host, vs.Namespace)
			if err != nil {
				fmt.Printf("Unable to convert hostname: %s\n", err.Error())
				return nil, err
			}

			// One VS can have multiple hosts and gateways. Make 1 key per
			// combination.
			hg := hostAndGateway{hostname: h}
			if len(vs.Spec.GetGateways()) > 0 {
				for _, g := range vs.Spec.GetGateways() {
					hg.gateway = g
					populateVirtualServiceMap(hg, vs, vsByHostAndGateway)
				}
			} else {
				hg.gateway = defaultGateway
				populateVirtualServiceMap(hg, vs, vsByHostAndGateway)
			}
		}
	}

	// create vet notes
	noteSet := map[conflictingVsNote][]string{}
	notes := []*apiv1.Note{}
	for key, vsList := range vsByHostAndGateway {
		if len(vsList) > 1 {
			conflictingRules := validateMergedVirtualServices(vsList)
			for _, conflict := range conflictingRules {
				vs1 := conflict[0]
				vs2 := conflict[1]
				vsNames := []string{vs1.vsName + "." + vs1.namespace, vs2.vsName + "." + vs2.namespace}
				conflictingRoutes := []string{vs1.route + " " + asString(vs1.ruleType),
					vs2.route + " " + asString(vs2.ruleType)}
				note := conflictingVsNote{
					Type:    vsHostNoteType,
					Summary: vsHostSummary,
					Msg:     vsHostMsg,
					Level:   apiv1.NoteLevel_ERROR,
					vsNames: strings.Join(vsNames, ", "),
					gateway: key.gateway,
					routes:  strings.Join(conflictingRoutes, " "),
				}
				noteSet[note] = append(noteSet[note], key.hostname)
			}
		}
	}
	for k, v := range noteSet {
		notes = append(notes, unwrapNote(k, v))
	}
	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}
	return notes, nil
}

func validateMergedVirtualServices(vsList []*v1alpha3.VirtualService) [][]routeRule {
	trie := buildMergedVirtualServiceTrie(vsList)
	// Do not try to validate when there is more than one regex.
	// Ideally, we should warn when there is more than one (since determining
	// whether one regex conflicts with another is very difficult), but this
	// should go in another vetter since reporting that error here would be
	// awkward and expands the scope of this vetter.
	if len(trie.regexs) == 1 {
		return validateVsTrie(trie, trie.regexs[0])
	} else {
		return validateVsTrie(trie, routeRule{})
	}
}

func buildMergedVirtualServiceTrie(vsList []*v1alpha3.VirtualService) *routeTrie {
	subRoutes := make(map[string]*routeTrie)
	trie := &routeTrie{subRoutes: subRoutes, regexs: []routeRule{}, routeRules: []routeRule{}}
	for _, vs := range vsList {
		for _, route := range vs.Spec.GetHttp() {
			for _, match := range route.GetMatch() {
				addRouteToMergedVsTree(trie, match.GetUri(), vs)
			}
		}
	}
	return trie
}

func addRouteToMergedVsTree(trie *routeTrie, match *istiov1alpha3.StringMatch, vs *v1alpha3.VirtualService) {
	current := trie
	rRule := getRouteRuleFromMatch(match, vs)

	// Regexs are treated as exceptions to the trie construction rule.
	// This is largely due to the complexities in determining whether two regexs
	// conflict.
	if rRule.ruleType == regex {
		trie.regexs = append(trie.regexs, rRule)
		return
	}

	if strings.HasSuffix("/", rRule.route) {
		rRule.route = strings.TrimSuffix(rRule.route, "/")
	}

	// Routes have leading slashes, remove the leading empty string from the array after the split
	components := strings.Split(rRule.route, "/")[1:]
	for count, component := range components {
		if next, ok := current.subRoutes[component]; ok {
			if count == len(components)-1 {
				next.routeRules = append(next.routeRules, rRule)
			} else {
				current = next
			}
		} else {
			newSubRoutes := make(map[string]*routeTrie)
			// This is the final component in a route rule and a new node is created for it
			if count == len(components)-1 {
				current.subRoutes[component] = &routeTrie{subRoutes: newSubRoutes, routeRules: []routeRule{rRule}}
			} else {
				newSubRoute := &routeTrie{subRoutes: newSubRoutes, routeRules: []routeRule{}}
				current.subRoutes[component] = newSubRoute
				current = newSubRoute
			}
		}
	}
}

func validateVsTrie(trie *routeTrie, rRule routeRule) [][]routeRule {
	conflictingRules := [][]routeRule{}
	for _, rule := range trie.routeRules {
		if conflict(rRule, rule) {
			conflictingRules = append(conflictingRules, []routeRule{rRule, rule})
		}
	}

	for _, descendant := range trie.subRoutes {
		if len(descendant.routeRules) == 0 {
			conflictingRules = append(conflictingRules, validateVsTrie(descendant, rRule)...)
		} else {
			// Recurse down but carefully! We want to report all conflicts and
			// we'll skip potential conflicts with the current route rule if we
			// recurse in the previous for loop (with the descendant rule as the "rRule" variable),
			for idx, rule := range append(descendant.routeRules, rRule) {
				if conflict(rRule, rule) {
					conflictingRules = append(conflictingRules, []routeRule{rRule, rule})
				}
				if idx < len(descendant.routeRules) {
					// remove "rule" to prevent double counting of conflicts
					//
					// This block checks which of the rules defined in a same route conflict and recurses
					// down with the given rule.
					newRouteRules := descendant.routeRules[idx+1:]
					newDescendant := &routeTrie{subRoutes: descendant.subRoutes, routeRules: newRouteRules}
					conflictingRules = append(conflictingRules, validateVsTrie(newDescendant, rule)...)
				} else {
					conflictingRules = append(conflictingRules, validateVsTrie(descendant, rule)...)
				}
			}
		}
	}
	return conflictingRules
}

// Document what's going on here better. Break down the cases, etc.,
func conflict(ancestorRule routeRule, descendantRule routeRule) bool {
	// The "(routeRule{})" needs to be in parenthesis; I'm not sure why.
	if ancestorRule == (routeRule{}) {
		return false
	}

	if ancestorRule.vsName == descendantRule.vsName {
		if ancestorRule == descendantRule {
			return false
		}
		if ancestorRule.ruleType == prefix {
			return strings.HasPrefix(descendantRule.route, ancestorRule.route)
		}
		if descendantRule.ruleType == prefix {
			return strings.HasPrefix(ancestorRule.route, descendantRule.route)
		}
		if ancestorRule != descendantRule && ancestorRule.ruleType == descendantRule.ruleType {
			return true
		} else {
			return false
		}
	} else {
		if ancestorRule.ruleType == regex {
			// Throwing away the error makes me feel gross but this regex should be validated before we get to it.
			// Even if it is invalid, giving the wrong answer isn't the biggest deal since regex validation is
			// outside the scope of this vetter.
			matched, _ := regexp.MatchString(ancestorRule.route, descendantRule.route)
			return matched
		}
		// Two routes in different virtual services with the same route are in conflict.
		if ancestorRule.route == descendantRule.route {
			return true
		} else if ancestorRule.ruleType == exact {
			if ancestorRule.route == descendantRule.route {
				return true
			} else {
				return false
			}
		} else {
			// Then ancestorRule is a prefix rule and, if descendant rule starts with that prefix,
			// it's in conflict.
			return strings.HasPrefix(descendantRule.route, ancestorRule.route)
		}
	}
}

func getRouteRuleFromMatch(match *istiov1alpha3.StringMatch, vs *v1alpha3.VirtualService) routeRule {
	if route := match.GetExact(); route != "" {
		return routeRule{ruleType: exact, route: route, vsName: vs.Name, namespace: vs.Namespace}
	} else if route := match.GetPrefix(); route != "" {
		return routeRule{ruleType: prefix, route: route, vsName: vs.Name, namespace: vs.Namespace}
	} else if route := match.GetRegex(); route != "" {
		return routeRule{ruleType: regex, route: route, vsName: vs.Name, namespace: vs.Namespace}
	}
	return routeRule{}
}

func populateVirtualServiceMap(hg hostAndGateway, vs *v1alpha3.VirtualService, vsByHostAndGateway VirtualSvcByHostAndGateway) {
	if _, ok := vsByHostAndGateway[hg]; !ok {
		vsByHostAndGateway[hg] = []*v1alpha3.VirtualService{vs}
	} else {
		vsByHostAndGateway[hg] = append(vsByHostAndGateway[hg], vs)
	}
}

// Vet returns the list of generated notes
func (v *VsHost) Vet() ([]*apiv1.Note, error) {
	virtualServices, err := util.ListVirtualServicesInMesh(v.nsLister, v.vsLister)
	if err != nil {
		fmt.Printf("Error occurred retrieving VirtualServices: %s\n", err.Error())
		return nil, err
	}
	notes, err := CreateVirtualServiceNotes(virtualServices)
	if err != nil {
		fmt.Printf("Error creating Conflicting VirtualService notes: %s\n", err.Error())
		return nil, err
	}
	return notes, nil
}

// Info returns information about the vetter
func (v *VsHost) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "VsHost" which implements the Vetter Tnterface
func NewVetter(factory vetter.ResourceListGetter) *VsHost {
	return &VsHost{
		nsLister: factory.K8s().Core().V1().Namespaces().Lister(),
		vsLister: factory.Istio().Networking().V1alpha3().VirtualServices().Lister(),
	}
}

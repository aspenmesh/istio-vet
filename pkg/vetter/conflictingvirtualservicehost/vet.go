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
	vsHostSummary  = "Multiple VirtualServices define the same host (${host}) and conflict"
	vsHostMsg      = "The VirtualServices ${vs_names} matching uris ${routes}" +
		" define the same host (${host}) and conflict. VirtualServices defining the same host must" +
		" not conflict. Consider updating the VirtualServices to have uniques hostname or " +
		"update the rules so they do not conflict."

	sidecarRoutingSummary = "Multiple VirtualServices define the same host and ${vs_name} uses sidecar routing."
	sidecarRoutingMsg     = "The VirtualService ${vs_name} uses sidecar routing (no gateway is explicitly attached)." +
		" However, other VirtualServices use " +
		"the same host: ${vs_names}, either fix the conflicts or add a non-default gateway to ${vs_name}"
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
			"routes":   note.routes,
		}}

}

// CreateVirtualServiceNotes checks for multiple vs defining the same host and
// generates notes for these cases
func CreateVirtualServiceNotes(virtualServices []*v1alpha3.VirtualService) ([]*apiv1.Note, error) {
	vsByHost := map[string][]*v1alpha3.VirtualService{}
	for _, vs := range virtualServices {
		for _, host := range vs.Spec.GetHosts() {
			h, err := util.ConvertHostnameToFQDN(host, vs.Namespace)
			if err != nil {
				fmt.Printf("Unable to convert hostname: %s\n", err.Error())
				return nil, err
			}
			if _, ok := vsByHost[h]; !ok {
				vsByHost[h] = []*v1alpha3.VirtualService{vs}
			} else {
				vsByHost[h] = append(vsByHost[h], vs)
			}
		}
	}

	// create vet notes

	// We only want to report the unique hosts for a given conflict.
	// This should be thought of as a hash map from notes to a set of
	// host names.
	notesToUniqueHost := map[conflictingVsNote]map[string]struct{}{}
	notes := []*apiv1.Note{}
	for host, vsList := range vsByHost {
		if len(vsList) > 1 {
			conflictingRules, err := conflictingVirtualServices(vsList)
			if err != nil {
				return notes, err
			}
			for _, vs := range vsList {
				gateways := vs.Spec.GetGateways()
				// If there are no gateways, then
				// this vs uses sidecar routing.
				// Since there are multiple virtual
				// services using the same host,
				// merging needs to occur, which
				// cannot happen with sidecar routing.
				if len(gateways) == 0 {
					vsNames := ""
					for _, sidecarRoutingVs := range vsList {
						if vsNames == "" {
							vsNames = sidecarRoutingVs.GetName()
						} else {
							vsNames = vsNames + ", " + sidecarRoutingVs.GetName()
						}
					}
					notes = append(notes, &apiv1.Note{
						Type:    vsHostNoteType,
						Summary: sidecarRoutingSummary,
						Msg:     sidecarRoutingMsg,
						Level:   apiv1.NoteLevel_ERROR,
						Attr: map[string]string{
							"vs_name":  vs.GetName(),
							"vs_names": vsNames,
						},
					})
				}
			}
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
					routes:  strings.Join(conflictingRoutes, " "),
				}
				notesToUniqueHost[note] = map[string]struct{}{host: struct{}{}}
			}
		}
	}
	for k, v := range notesToUniqueHost {
		hosts := []string{}
		for host, _ := range v {
			hosts = append(hosts, host)
		}
		notes = append(notes, unwrapNote(k, hosts))
	}
	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}
	return notes, nil
}

// Return a list of pairs of virtual services that conflict.
func conflictingVirtualServices(vsList []*v1alpha3.VirtualService) ([][]routeRule, error) {
	trie := buildMergedVirtualServiceTrie(vsList)
	// Do not try to validate when there is more than one regex.
	// Ideally, we should warn when there is more than one (since determining
	// whether one regex conflicts with another is very difficult), but this
	// should go in another vetter since reporting that error here would be
	// awkward and expands the scope of this vetter.
	conflictingRules := addConflictsForSameRoute(trie, [][]routeRule{})
	if len(trie.regexs) == 1 {
		if rules, err := conflictingSubroutes(trie, trie.regexs[0], conflictingRules); err != nil {
			return [][]routeRule{}, err
		} else {
			return rules, nil
		}
	} else {
		if rules, err := conflictingSubroutes(trie, routeRule{}, conflictingRules); err != nil {
			return [][]routeRule{}, err
		} else {
			return rules, nil
		}
	}
}

// Create a trie representing the routes with their corresponding match rules
// from a list of virtual services.
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

// Add a particular route to the route trie. If the given route already has a route rule,
// add it to the list of route rules for the given node/route.
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

func conflictingSubroutes(trie *routeTrie, rRule routeRule, conflictingRules [][]routeRule) ([][]routeRule, error) {
	for _, rule := range trie.routeRules {
		if c, err := conflict(rRule, rule); err != nil {
			return conflictingRules, err
		} else {
			if c {
				conflictingRules = append(conflictingRules, []routeRule{rRule, rule})
			}
		}
	}

	for _, descendant := range trie.subRoutes {
		if len(descendant.routeRules) == 0 {
			if c, err := conflictingSubroutes(descendant, rRule, conflictingRules); err != nil {
				return conflictingRules, err
			} else {
				conflictingRules = append(conflictingRules, c...)
			}
		} else {
			// Recurse down but carefully! We want to report all conflicts and
			// we'll skip potential conflicts with the current route rule if we
			// recurse in the previous for loop (with the descendant rule as the "rRule" variable),
			for _, rule := range append(trie.routeRules, rRule) {
				newRules, err := conflictingSubroutes(descendant, rule, conflictingRules)
				if err != nil {
					return conflictingRules, err
				}
				conflictingRules = newRules
			}
		}
	}
	return conflictingRules, nil
}
func addConflictsForSameRoute(trie *routeTrie, conflictingRules [][]routeRule) [][]routeRule {
	routeRules := trie.routeRules
	for i := 0; i < len(routeRules)-1; i++ {
		for j := i + 1; j < len(routeRules); j++ {
			conflictingRules = append(conflictingRules, []routeRule{routeRules[i], routeRules[j]})
		}
	}

	for _, descendant := range trie.subRoutes {
		return addConflictsForSameRoute(descendant, conflictingRules)
	}
	return conflictingRules
}

// Returns true if the rules conflict, false otherwise.
//
// There are several cases that we need to keep track of:
//
// case 1: ancestor rule is an empty struct
//   This happens after we've encountered our first "real" route rule. This should
//   always return false.
//
// case 2: There is exactly one regex in the trie
//   For reasons elaborated elsewhere, there will only ever be one regex in
//   a given trie. Given how the trie is traversed, the regex will always be
//   the ancestorRule; it can never be the descendantRule.
//
// case 3: The ancestorRule is a prefix rule
//   Note that the only relevant case here is when the route for descendantRule is a
//   strict subroute of ancestorRule (because rules for the same route are handled in a different
//   code path). In practice, this should always result in a conflict because of how the trie is traversed.
//
//  case 4: The ancestorRule is an exact rule
//    Since the only relevant case is when the route for descendantRule is a strict
//    subroute of ancestorRule (same as case 3), this should always be false in practice.
func conflict(ancestorRule routeRule, descendantRule routeRule) (bool, error) {
	// The "(routeRule{})" needs to be in parenthesis; I'm not sure why.
	if ancestorRule == (routeRule{}) {
		return false, nil
	}

	if ancestorRule.ruleType == regex {
		matched, err := regexp.MatchString(ancestorRule.route, descendantRule.route)
		return matched, err
	}

	if ancestorRule.ruleType == prefix {
		// This should always be true (since a given rule will only check its "descendants"),
		// but this is more explicit.
		return strings.HasPrefix(descendantRule.route, ancestorRule.route), nil
	}

	if ancestorRule.ruleType == exact {
		// Since a rule will only be checked against its strict descendants or a rule on the same
		// path (which is checked earlier), this should always be false. This is more explicit, though.
		return ancestorRule.route == descendantRule.route, nil
	}

	return true, fmt.Errorf("Could not determine whether these %v and %v are in conflict! This "+
		"is the result of a bug in the vetter.", ancestorRule, descendantRule)
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

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
		" not conflict. Consider updating the VirtualServices to have unique hostnames or " +
		"update the rules so they do not conflict."
)

type routeRuleType int

const (
	prefix routeRuleType = iota
	exact
	regex
)

// VsHost implements Vetter interface
type VsHost struct {
	nsLister v1.NamespaceLister
	vsLister netv1alpha3.VirtualServiceLister
}

type routeRule struct {
	ruleType  routeRuleType
	route     string
	vsName    string
	namespace string
	priority  int
}

type routeTrie struct {
	subRoutes  map[string]*routeTrie
	regexs     []routeRule
	routeRules []routeRule
}

func asString(rrType routeRuleType) string {
	if rrType == prefix {
		return "prefix"
	} else if rrType == exact {
		return "exact"
	} else {
		return "regex"
	}
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
	notes, err := addConflictingRulesNotes(vsByHost)

	if err != nil {
		return []*apiv1.Note{}, err
	}
	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}
	return notes, nil
}

func addConflictingRulesNotes(vsByHost map[string][]*v1alpha3.VirtualService) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	for host, vsList := range vsByHost {
		if len(vsList) >= 1 {

			conflictingRules, err := conflictingVirtualServices(vsList)
			if err != nil {
				return notes, err
			}
			for _, conflict := range conflictingRules {
				vs1 := conflict[0]
				vs2 := conflict[1]
				vsNames := []string{vs1.vsName + "." + vs1.namespace, vs2.vsName + "." + vs2.namespace}
				conflictingRoutes := []string{vs1.route + " " + asString(vs1.ruleType),
					vs2.route + " " + asString(vs2.ruleType)}
				note := &apiv1.Note{
					Type:    vsHostNoteType,
					Summary: vsHostSummary,
					Msg:     vsHostMsg,
					Level:   apiv1.NoteLevel_ERROR,
					Attr: map[string]string{
						"vs_names": strings.Join(vsNames, ", "),
						"host":     host,
						"routes":   strings.Join(conflictingRoutes, " "),
					},
				}
				notes = append(notes, note)
			}
		}
	}

	return notes, nil
}

// Return a list of pairs of virtual services that conflict.
func conflictingVirtualServices(vsList []*v1alpha3.VirtualService) ([][]routeRule, error) {
	trie := buildMergedVirtualServiceTrie(vsList)
	conflictingRules := addConflictsForSameRoute(trie, [][]routeRule{})

	// Do not try to validate when there is more than one regex.
	// Ideally, we should warn when there is more than one (since determining
	// whether one regex conflicts with another is computationally very difficult), but this
	// should go in another vetter since reporting that error here would be
	// awkward and expands the scope of this vetter.
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
//
// The nodes of the trie can be thought of as the components of a route with
// arrays/slices containing the match type of the route rules for that node
// if there are any.
//
// For example, consider the route rules:
//   /foo/bar exact
//   /foo/bar prefix
//   /bar exact
//   /bar/baz prefix
//
// Then the trie could be thought of like
//
//                         o (dummy node)
//                        / \
//                       /   \
//                      /     \
//                     /       \
//                (/foo, [])  (/bar, [exact])
//                   /          \
//                  /            \
//                 /              \
// (/foo/bar, [prefix, exact])    (/bar/baz, [prefix])
func buildMergedVirtualServiceTrie(vsList []*v1alpha3.VirtualService) *routeTrie {
	subRoutes := make(map[string]*routeTrie)
	trie := &routeTrie{subRoutes: subRoutes, regexs: []routeRule{}, routeRules: []routeRule{}}
	for _, vs := range vsList {
		for prio, route := range vs.Spec.GetHttp() {
			for _, match := range route.GetMatch() {
				addRouteToMergedVsTree(trie, match.GetUri(), vs, prio)
			}
		}
	}
	return trie
}

// Add a particular route to the route trie. If the given route already has a route rule,
// add it to the list of route rules for the given node/route.
func addRouteToMergedVsTree(trie *routeTrie, match *istiov1alpha3.StringMatch, vs *v1alpha3.VirtualService, prio int) {
	current := trie
	rRule := getRouteRuleFromMatch(match, vs, prio)

	// Regexs are treated as exceptions to the trie construction rule.
	// This is largely due to the complexities in determining whether two regexs
	// conflict.
	if rRule.ruleType == regex {
		trie.regexs = append(trie.regexs, rRule)
		return
	}

	// Trim trailing slashes
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

// Traverse the trie depth-first and add any conflicts to the list of conflicting rules.
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
		// There are no route rules for this node, recurse down with the same rule.
		if len(descendant.routeRules) == 0 {
			if c, err := conflictingSubroutes(descendant, rRule, conflictingRules); err != nil {
				return conflictingRules, err
			} else {
				conflictingRules = c
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

// Add conflicts for the same route to the list of conflicting rules. Traverse the trie depth-first.
// Rules for a given route will always conflict if they are not in the same virtual service.
// At root, routeRules should be all regex, therefore, we should skip check of the regex rules
func addConflictsForSameRoute(trie *routeTrie, conflictingRules [][]routeRule) [][]routeRule {
	routeRules := trie.routeRules
	for i := 0; i < len(routeRules)-1; i++ {
		for j := i + 1; j < len(routeRules); j++ {
			// Do not report when the rules are in the same virtual service.
			// Order matters for conflicting rules in the same virtual service,
			// however, this can be finnicky enough that I'm leaving it out
			// of the first pass and we can add it in later if it is a cause
			// of confusion.
			if routeRules[i].vsName != routeRules[j].vsName ||
				routeRules[i].namespace != routeRules[j].namespace {
				conflictingRules = append(conflictingRules, []routeRule{routeRules[i], routeRules[j]})
			} else if routeRules[i].vsName == routeRules[j].vsName {
				if c, err := sameVSconflict(routeRules[i], routeRules[j]); err == nil && c {
					conflictingRules = append(conflictingRules, []routeRule{routeRules[i], routeRules[j]})
				}
		    }
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
// case 2: Ancestor and descendant are in the same virtual service:
//   Order of declaration matters when applying route rules from the
//   same virtual service. However, the current implementation doesn't
//   track this. Always report no conflict for now.
//
// case 3: There is exactly one regex in the trie
//   For reasons elaborated elsewhere, there will only ever be one regex in
//   a given trie. Given how the trie is traversed, the regex will always be
//   the ancestorRule; it can never be the descendantRule.
//
// case 4: The ancestorRule is a prefix rule
//   The only relevant case here is when the route for descendantRule is a
//   strict subroute of ancestorRule (because rules for the same route are handled in a different
//   code path). This should always return true because of how the trie is traversed.
//
// case 5: The ancestorRule is an exact rule
//   Since the only relevant case is when the route for descendantRule is a strict
//   subroute of ancestorRule (same as case 3), this should always be false.
func conflict(ancestorRule routeRule, descendantRule routeRule) (bool, error) {
	// The "(routeRule{})" needs to be in parenthesis; I'm not sure why.
	if ancestorRule == (routeRule{}) {
		return false, nil
	}

	if ancestorRule.vsName == descendantRule.vsName && ancestorRule.namespace == descendantRule.namespace {
		return sameVSconflict(ancestorRule, descendantRule)
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

func getRouteRuleFromMatch(match *istiov1alpha3.StringMatch, vs *v1alpha3.VirtualService, prio int) routeRule {
	if route := match.GetExact(); route != "" {
		return routeRule{ruleType: exact, route: route, vsName: vs.Name, namespace: vs.Namespace, priority: prio}
	} else if route := match.GetPrefix(); route != "" {
		return routeRule{ruleType: prefix, route: route, vsName: vs.Name, namespace: vs.Namespace, priority: prio}
	} else if route := match.GetRegex(); route != "" {
		return routeRule{ruleType: regex, route: route, vsName: vs.Name, namespace: vs.Namespace, priority: prio}
	}
	return routeRule{}
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

// check same VS route conflict
func sameVSconflict(rule1 routeRule, rule2 routeRule) (bool, error) {

       if (rule1.priority > rule2.priority &&
		((rule1.ruleType == prefix && (rule2.ruleType == prefix || rule2.ruleType == exact)) ||
			(rule1.ruleType == exact && rule2.ruleType == exact))) {
		return strings.HasPrefix(rule2.route, rule1.route), nil

	} else if (rule2.priority > rule1.priority &&
		((rule2.ruleType == prefix && (rule1.ruleType == prefix || rule1.ruleType == exact)) ||
			   (rule2.ruleType == exact && rule1.ruleType == exact))) {
		return strings.HasPrefix(rule1.route, rule2.route), nil
	}
	return false, nil
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

/*
Copyright 2017 Aspen Mesh Authors.

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
package authpolicy

import (
	"fmt"
	"sort"
	"strings"

	aspenv1a1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	v1alpha1 "github.com/aspenmesh/istio-client-go/pkg/client/listers/authentication/v1alpha1"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	vetterID                    = "AuthPolicyConflict"
	authPolicySummary           = "Conflicting authentication policies - ${policy_names}"
	authPolicyNoteTypeNamespace = "auth-policy-conflict-namespace"
	authPolicyNoteTypeService   = "auth-policy-conflict-service"
	authPolicyNoteTypePorts     = "auth-policy-conflict-port"
	authPolicyNamespaceMsg      = "Multiple authentication policies (${policy_names}) in namespace ${namespace} set the namespace-wide config which will cause unwanted behavior. Update policies to remove conflicts."
	authPolicyTargetSvcNameMsg  = "Multiple authentication policies (${policy_names}) in namespace ${namespace} set the service-wide config for ${target_service} which will cause unwanted behavior. Update policies to remove conflicts."
	authPolicySvcPortMsg        = "Multiple authentication policies (${policy_names}) in namespace ${namespace} sets the service port config for ${target_service}:${target_port} which will cause unwanted behavior. Update policies to remove conflicts."
)

type AuthPolicy struct {
	polLister v1alpha1.PolicyLister
}

type PolicyKey struct {
	Namespace  string
	TargetName string
	TargetPort uint32
}

func notesForAuthPolicies(policies []*aspenv1a1.Policy) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	polKeyMap := map[PolicyKey][]*aspenv1a1.Policy{}

	if len(policies) == 0 {
		return notes, nil
	}

	for p := range policies {
		targets := policies[p].Spec.Policy.Targets

		if len(targets) > 0 {
			for t := range targets {
				targetPort := targets[t].Ports
				if len(targetPort) > 0 {
					// iterate through each target in the list and make a new key
					for o := range targetPort {
						//if there is a port, iterate through targetPort
						pk := PolicyKey{
							Namespace:  policies[p].Namespace,
							TargetName: targets[t].Name,
							TargetPort: targetPort[o].GetNumber(),
						}
						polKeyMap[pk] = append(polKeyMap[pk], policies[p])
					}
				} else {
					//just make the one key if there is no port
					pk := PolicyKey{
						Namespace:  policies[p].Namespace,
						TargetName: targets[t].Name,
						TargetPort: 0,
					}
					polKeyMap[pk] = append(polKeyMap[pk], policies[p])
				}
			}
		} else {
			pk := PolicyKey{
				Namespace:  policies[p].Namespace,
				TargetName: "",
				TargetPort: 0,
			}
			polKeyMap[pk] = append(polKeyMap[pk], policies[p])
		}

	}

	for k, policyList := range polKeyMap {
		if len(policyList) > 1 {
			policyNames := []string{}
			for p := range policyList {
				policyNames = append(policyNames, policyList[p].ObjectMeta.Name)
			}
			sort.Slice(policyNames, func(i, j int) bool {
				return policyNames[i] < policyNames[j]
			})

			if k.TargetPort == 0 && k.TargetName == "" {
				// policy conflict at namespace level
				notes = append(notes, &apiv1.Note{
					Type:    authPolicyNoteTypeNamespace,
					Summary: authPolicySummary,
					Msg:     authPolicyNamespaceMsg,
					Level:   apiv1.NoteLevel_ERROR,
					Attr: map[string]string{
						"namespace":    k.Namespace,
						"policy_names": strings.Join(policyNames, ", ")}})

			} else if k.TargetPort == 0 {
				// policy conflict of duplicated Service Names
				notes = append(notes, &apiv1.Note{
					Type:    authPolicyNoteTypeService,
					Summary: authPolicySummary,
					Msg:     authPolicyTargetSvcNameMsg,
					Level:   apiv1.NoteLevel_ERROR,
					Attr: map[string]string{
						"namespace":      k.Namespace,
						"policy_names":   strings.Join(policyNames, ", "),
						"target_service": k.TargetName}})
			} else {
				// policy conflict at namespace level
				notes = append(notes, &apiv1.Note{
					Type:    authPolicyNoteTypePorts,
					Summary: authPolicySummary,
					Msg:     authPolicySvcPortMsg,
					Level:   apiv1.NoteLevel_ERROR,
					Attr: map[string]string{
						"namespace":      k.Namespace,
						"policy_names":   strings.Join(policyNames, ", "),
						"target_service": k.TargetName,
						"target_port":    fmt.Sprintf("%v", k.TargetPort)}})
			}
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

func (a *AuthPolicy) Vet() ([]*apiv1.Note, error) {
	policies, err := a.polLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	notes, err := notesForAuthPolicies(policies)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// Info returns information about the vetter
func (a *AuthPolicy) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

//NewVetter returns "AuthPolicy" which implements Vetter interface
func NewVetter(factory vetter.ResourceListGetter) *AuthPolicy {
	return &AuthPolicy{
		polLister: factory.Istio().Authentication().V1alpha1().Policies().Lister(),
	}
}

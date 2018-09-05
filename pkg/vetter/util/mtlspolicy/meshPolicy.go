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

func IsGlobalMtlsEnabled(meshPolicies []*authv1alpha1.MeshPolicy) (bool, error) {
	if len(meshPolicies) > 1 {
		return false, errors.New("More than one MeshPolicy was found")
	} else if len(meshPolicies) == 0 {
		return false, nil
	} else {
		if strings.EqualFold(meshPolicies[0].ObjectMeta.Name, "default") {
			return MeshPolicyIsMtls(meshPolicies[0]), nil
		}
		return false, errors.New("MeshPolicy is not named 'default'")
	}
}

// AuthPolicyIsMtls returns true if the passed Policy has mTLS enabled
func MeshPolicyIsMtls(meshPolicy *authv1alpha1.MeshPolicy) bool {
	peers := meshPolicy.Spec.GetPeers()
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

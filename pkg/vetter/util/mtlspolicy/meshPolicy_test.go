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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authv1alpha1 "github.com/aspenmesh/istio-client-go/pkg/apis/authentication/v1alpha1"
	istioauthv1alpha1 "istio.io/api/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type expectedValue struct {
	input       []*authv1alpha1.MeshPolicy
	mtlsEnabled bool
	err         error
}

var (
	defaultOn = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}

	defaultOff = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{},
			},
		},
	}

	nonDefaultNamedOn = &authv1alpha1.MeshPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MeshPolicy",
			APIVersion: "authentication.istio.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "named",
		},
		Spec: authv1alpha1.MeshPolicySpec{
			Policy: istioauthv1alpha1.Policy{
				Peers: []*istioauthv1alpha1.PeerAuthenticationMethod{
					&istioauthv1alpha1.PeerAuthenticationMethod{
						Params: &istioauthv1alpha1.PeerAuthenticationMethod_Mtls{},
					},
				},
			},
		},
	}
)

var _ = Describe("Test the status of mTLS from MeshPolicies", func() {
	var expectedValues = []expectedValue{
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{},
			mtlsEnabled: false,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOn, defaultOff},
			mtlsEnabled: false,
			err:         errors.New(""),
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOn},
			mtlsEnabled: true,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{defaultOff},
			mtlsEnabled: false,
			err:         nil,
		},
		expectedValue{
			input:       []*authv1alpha1.MeshPolicy{nonDefaultNamedOn},
			mtlsEnabled: false,
			err:         errors.New(""),
		},
	}
	It("Should match the expected values", func() {
		for i := 0; i < len(expectedValues); i++ {
			mtlsEnabled, err := IsGlobalMtlsEnabled(expectedValues[i].input)
			if expectedValues[i].err != nil {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(mtlsEnabled).To(Equal(expectedValues[i].mtlsEnabled))
		}
	})

})

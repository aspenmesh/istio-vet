/*
Copyright 2019 Aspen Mesh Authors.

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

package services

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceFromFqdn", func() {
	It("should accept valid FQDNs", func() {
		type passTc struct {
			Fqdn    string
			Service Service
		}
		passTestCases := []passTc{
			passTc{
				Fqdn:    "foo.default.svc.cluster.local",
				Service: Service{Name: "foo", Namespace: "default"},
			},
			passTc{
				Fqdn:    "bar.svc.svc.cluster.local",
				Service: Service{Name: "bar", Namespace: "svc"},
			},
		}
		for _, testCase := range passTestCases {
			s, err := ServiceFromFqdn(testCase.Fqdn)
			Expect(err).To(Succeed())
			if err != nil {
				continue
			}
			Expect(s.Name).To(Equal(testCase.Service.Name))
			Expect(s.Namespace).To(Equal(testCase.Service.Namespace))
		}
	})
	It("should reject invalid FQDNs", func() {
		failTestCases := []string{
			"cluster.local",
			"svc.cluster.local",
			"default.svc.cluster.local",
			".default.svc.cluster.local",
			"..default.svc.cluster.local",
			"foo.default.svc.cluster.local.",
		}

		for _, testCase := range failTestCases {
			_, err := ServiceFromFqdn(testCase)
			Expect(err).To(HaveOccurred())
		}
	})
})

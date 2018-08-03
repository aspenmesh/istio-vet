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

package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converting short hostnames to FQDN", func() {
	Context("In namespace 'foo'", func() {
		namespace := "foo"

		It("Returns an error when hostname and/or namespace are passed as empty strings", func() {
			_, err := ConvertHostnameToFQDN("", namespace)
			Expect(err).To(HaveOccurred())
			_, err = ConvertHostnameToFQDN("", "")
			Expect(err).To(HaveOccurred())
			_, err = ConvertHostnameToFQDN("host", "")
			Expect(err).To(HaveOccurred())
		})

		It("Returns a FQDN when given a short host name", func() {
			givenHostname := "reviews"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			expectedFqdnHostname := givenHostname + "." + namespace + ".svc.cluster.local"
			Expect(returnedHostname).NotTo(Equal(givenHostname))
			Expect(returnedHostname).To(Equal(expectedFqdnHostname))
		})

		It("Does not return new host name when given a FQDN", func() {
			givenHostname := "reviews.foo.svc.cluster.local"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedHostname).To(Equal(givenHostname))
		})

		It("Does not return a new host name when given an IP address", func() {
			givenHostname := "255.255.255.0"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedHostname).To(Equal(givenHostname))
		})

		It("Does not return a new host name when given an *", func() {
			givenHostname := "*"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedHostname).To(Equal(givenHostname))
		})

		It("Does not return a new host name when given anything beginning with *", func() {
			givenHostname := "*.foo.com"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedHostname).To(Equal(givenHostname))
		})

		It("Does not return a new host name when given a web address", func() {
			givenHostname := "foo.com"
			returnedHostname, err := ConvertHostnameToFQDN(givenHostname, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedHostname).To(Equal(givenHostname))
		})
	})
})

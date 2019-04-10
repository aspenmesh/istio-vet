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
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
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

func configMapFromFile(file string) *corev1.ConfigMap {

	icm, err := ioutil.ReadFile(file)
	if err != nil {
		Fail(err.Error())
	}

	var configMap corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(icm), &configMap); err != nil {

		Fail(err.Error())
	}

	return &configMap
}

var _ = Describe("Converting configmap to SidecarInjectionSpec", func() {

	It("Can create a SidecarInjectionSpec from ConfigMaps", func() {
		file1 := "./testdata/0.8/0.8-istio-sidecar-injector.yaml"
		file2 := "./testdata/0.8/0.8-mesh-config.yaml"

		mockICM := configMapFromFile(file1)
		mockMCM := configMapFromFile(file2)

		sidecarInjSpec, err := makeSideCarSpec(mockICM, mockMCM)

		Expect(err).To(BeNil())
		Expect(sidecarInjSpec.InitContainers[0].Name).To(Equal("istio-init"))
		Expect(sidecarInjSpec.InitContainers[0].Image).To(Equal("docker.io/istio/proxy_init:0.8.0"))
		Expect(sidecarInjSpec.Containers[0].Name).To(Equal("istio-proxy"))
		Expect(sidecarInjSpec.Containers[0].Image).To(Equal("docker.io/istio/proxyv2:0.8.0"))
		Expect(sidecarInjSpec.Volumes[0].Name).To(Equal("istio-envoy"))

	})

})

var _ = Describe("Test ProxyStatusPort", func() {
	It("Finds an override that is not 15020", func() {

		// Typical argument list from a kubectl get pods
		testArgs := []string{"proxy",
			"sidecar",
			"--domain",
			"$(POD_NAMESPACE).svc.cluster.local",
			"--configPath",
			"/etc/istio/proxy",
			"--binaryPath",
			"/usr/local/bin/envoy",
			"--serviceCluster",
			"atings.$(POD_NAMESPACE)",
			"--drainDuration",
			"5s",
			"--parentShutdownDuration",
			"m0s",
			"--discoveryAddress",
			"istio-pilot.istio-system:15011",
			"--zipkinAddress",
			"zipkin.istio-system:9411",
			"--connectTimeout",
			"0s",
			"--proxyAdminPort",
			"15000",
			"--concurrency",
			"2",
			"--controlPlaneAuthPolicy",
			"UTUAL_TLS",
			"--statusPort",
			"15020",
			"--applicationPorts",
			"9080",
		}
		container := corev1.Container{Args: testArgs}
		port, err := ProxyStatusPort(container)
		Expect(port == 15020)
		Expect(err == nil)

		// Fail to parse a number value for statusPort
		container.Args = []string{"proxy",
			"sidecar",
			"--domain",
			"--statusPort",
			"junk",
		}
		port, err = ProxyStatusPort(container)
		Expect(port == kubernetesProxyStatusPortDefault)
		Expect(err != nil)

		// Status port is not the default
		container.Args = []string{"proxy",
			"sidecar",
			"--domain",
			"--statusPort",
			"12345",
		}
		port, err = ProxyStatusPort(container)
		Expect(port == 12345)
		Expect(err == nil)

		// No statusPort defined
		container.Args = []string{"proxy",
			"sidecar",
			"--domain",
		}
		port, err = ProxyStatusPort(container)
		Expect(port == kubernetesProxyStatusPortDefault)
		Expect(err != nil)

		// statusPort defined but not specified!
		container.Args = []string{"proxy",
			"sidecar",
			"--domain",
			"--statusPort",
		}
		port, err = ProxyStatusPort(container)
		Expect(port == kubernetesProxyStatusPortDefault)
		Expect(err != nil)
	})
})
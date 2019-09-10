/*
Portions Copyright 2017 Istio Authors
Portions Copyright 2017 Aspen Mesh Authors.

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

// Package util provides common constants and helper functions for vetters.
package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "github.com/aspenmesh/istio-client-go/pkg/client/listers/networking/v1alpha3"
	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/cnf/structhash"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	meshv1alpha1 "istio.io/api/mesh/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers/core/v1"
)

// Constants related to Istio
const (
	IstioNamespace                = "istio-system"
	IstioProxyContainerName       = "istio-proxy"
	IstioInitContainerName        = "istio-init"
	IstioConfigMap                = "istio"
	IstioConfigMapKey             = "mesh"
	IstioInitializerPodAnnotation = "sidecar.istio.io/status"
	IstioInitializerConfigMap     = "istio-sidecar-injector"
	IstioInitializerConfigMapKey  = "config"
	IstioAppLabel                 = "app"
	KubernetesDomainSuffix        = ".svc.cluster.local"
	ServiceProtocolUDP            = "UDP"
	initializerDisabled           = "configmaps \"" +
		IstioInitializerConfigMap + "\" not found"
	initializerDisabledSummary = "Istio initializer is not configured." +
		" Enable initializer and automatic sidecar injection to use "
	kubernetesServiceName            = "kubernetes"
	kubernetesProxyStatusPort        = "--statusPort"
	kubernetesProxyStatusPortDefault = 15020
)

var istioInjectNamespaceLabel = map[string]string{
	"istio-injection": "enabled"}

// Config specifies the sidecar injection configuration This includes
// the sidear template and cluster-side injection policy. It is used
// by kube-inject, sidecar injector, and http endpoint.
type IstioInjectConfig struct {
	Policy InjectionPolicy `json:"policy"`

	// Template is the templated version of `SidecarInjectionSpec` prior to
	// expansion over the `SidecarTemplateData`.
	Template string `json:"template"`
}

var istioSupportedServicePrefix = []string{
	"http", "http-",
	"http2", "http2-",
	"https", "https-",
	"grpc", "grpc-",
	"mongo", "mongo-",
	"redis", "redis-",
	"tcp", "tcp-",
	"tls", "tls-",
	"udp", "udp-"}

var defaultExemptedNamespaces = map[string]bool{
	"kube-system":  true,
	"kube-public":  true,
	"istio-system": true}

// DefaultExemptedNamespaces returns list of default Namsepaces which are
// exempted from automatic sidecar injection.
// List includes "kube-system", "kube-public" and "istio-system"
func DefaultExemptedNamespaces() []string {
	s := make([]string, len(defaultExemptedNamespaces))
	i := 0
	for k := range defaultExemptedNamespaces {
		s[i] = k
		i++
	}
	return s
}

// ExemptedNamespace checks if a Namespace is by default exempted from automatic
// sidecar injection.
func ExemptedNamespace(ns string) bool {
	return defaultExemptedNamespaces[ns]
}

// GetInitializerConfig retrieves the Istio Initializer config.
// Istio Initializer config is stored as "istio-sidecar-injector" configmap in
// "istio-system" Namespace.
func GetInitializerConfigMap(cmLister v1.ConfigMapLister) (*corev1.ConfigMap, error) {
	cm, err := cmLister.ConfigMaps(IstioNamespace).Get(IstioInitializerConfigMap)
	if err != nil {
		glog.V(2).Infof("Failed to retrieve configmap: %s error: %s", IstioInitializerConfigMap, err)
		return nil, err
	}
	return cm, nil
}

// GetIstioInjectConfig is separated for testing in util_test.go
func GetIstioInjectConfig(cm *corev1.ConfigMap) (*IstioInjectConfig, error) {
	d, e := cm.Data[IstioInitializerConfigMapKey]
	if !e {
		errStr := fmt.Sprintf("Missing configuration map key: %s in configmap: %s", IstioInitializerConfigMapKey, IstioInitializerConfigMap)
		glog.Errorf(errStr)
		return nil, errors.New(errStr)
	}
	var cfg IstioInjectConfig
	if err := yaml.Unmarshal([]byte(d), &cfg); err != nil {
		glog.Errorf("Failed to parse yaml initializer config: %s", err)
		return nil, err
	}
	return &cfg, nil
}

// GetMeshConfig retrieves the Istio Mesh config.
// Istio Mesh config is stored as "istio" configmap in
// "istio-system" Namespace.
func GetMeshConfigMap(cmLister v1.ConfigMapLister) (*corev1.ConfigMap, error) {
	cm, err := cmLister.ConfigMaps(IstioNamespace).Get(IstioConfigMap)
	if err != nil {
		glog.Errorf("Failed to retrieve configmap: %s error: %s", IstioConfigMap, err)
		return nil, err
	}
	return cm, nil
}

func GetMeshConfig(cm *corev1.ConfigMap) (*meshv1alpha1.MeshConfig, error) {
	c, e := cm.Data[IstioConfigMapKey]
	if !e {
		errStr := fmt.Sprintf("Missing configuration map key: %s in configmap: %s", IstioConfigMapKey, IstioConfigMap)
		glog.Errorf(errStr)
		return nil, errors.New(errStr)
	}

	if len(c) == 0 {
		return nil, nil
	}
	var cfg meshv1alpha1.MeshConfig
	if err := ApplyYAML(c, &cfg, false); err != nil {
		glog.Errorf("Failed to parse yaml mesh config: %s", err)
		return nil, err
	}
	return &cfg, nil
}

func makeSideCarSpec(icm, mcm *corev1.ConfigMap) (*SidecarInjectionSpec, error) {
	ic, err := GetIstioInjectConfig(icm)
	if err != nil {
		return nil, err
	}
	mc, err := GetMeshConfig(mcm)
	if err != nil {
		return nil, err
	}
	version := "latest"
	spec, _, err := injectionData(ic.Template, version, &metav1.ObjectMeta{}, &corev1.PodSpec{}, &metav1.ObjectMeta{}, mc.DefaultConfig, mc)
	return spec, err
}

// GetInitializerSidecarSpec retrieves the sidecar spec which will be inserted
// by the initializer
func GetInitializerSidecarSpec(cmLister v1.ConfigMapLister) (*SidecarInjectionSpec, error) {

	configMap, err := GetInitializerConfigMap(cmLister)
	if err != nil {
		return nil, err
	}
	meshConfigMap, err := GetMeshConfigMap(cmLister)
	if err != nil {
		return nil, err
	}
	return makeSideCarSpec(configMap, meshConfigMap)
}

// IstioInitializerDisabledNote generates an INFO note if the error string
// contains "istio-inject configmap not found".
func IstioInitializerDisabledNote(e, vetterID, vetterType string) *apiv1.Note {
	if strings.Contains(e, initializerDisabled) {
		return &apiv1.Note{
			Type:    vetterType,
			Summary: initializerDisabledSummary + "\"" + vetterID + "\" vetter.",
			Level:   apiv1.NoteLevel_INFO}
	}
	return nil
}

// ServicePortPrefixed checks if the Service port name is prefixed with Istio
// supported protocols.
func ServicePortPrefixed(n string) bool {
	i := 0
	for i < len(istioSupportedServicePrefix) {
		if n == istioSupportedServicePrefix[i] || strings.HasPrefix(n, istioSupportedServicePrefix[i+1]) {
			return true
		}
		i += 2
	}
	return false
}

// SidecarInjected checks if sidecar is injected in a Pod.
// Sidecar is considered injected if initializer annotation and proxy container
// are both present in the Pod Spec.
func SidecarInjected(p *corev1.Pod) bool {
	if _, ok := p.Annotations[IstioInitializerPodAnnotation]; !ok {
		return false
	}
	cList := p.Spec.Containers
	for _, c := range cList {
		if c.Name == IstioProxyContainerName {
			return true
		}
	}
	return false
}

func imageFromContainers(n string, cList []corev1.Container) (string, error) {
	for _, c := range cList {
		if c.Name == n {
			return c.Image, nil
		}
	}
	errStr := fmt.Sprintf("Failed to find container %s", n)
	glog.Error(errStr)
	return "", errors.New(errStr)
}

// Image returns the image for the container named n if present
// in the pod spec, or an error otherwise.
func Image(n string, s corev1.PodSpec) (string, error) {
	return imageFromContainers(n, s.Containers)
}

// InitImage returns the image for the init container named n if present
// in the pod spec, or an error otherwise.
func InitImage(n string, s corev1.PodSpec) (string, error) {
	return imageFromContainers(n, s.InitContainers)
}

// ListNamespacesInMesh returns the list of Namespaces in the mesh.
// Namespaces with label "istio-inject=enabled" are considered in
// the mesh.
func ListNamespacesInMesh(nsLister v1.NamespaceLister) ([]*corev1.Namespace, error) {
	ns, err := nsLister.List(labels.Set(istioInjectNamespaceLabel).AsSelector())
	if err != nil {
		glog.Error("Failed to retrieve namespaces: ", err)
		return nil, err
	}
	return ns, nil
}

// ListPodsInMesh returns the list of Pods in the mesh.
// Pods in Namespaces returned by ListNamespacesInMesh with sidecar
// injected as determined by SidecarInjected are considered in the mesh.
func ListPodsInMesh(nsLister v1.NamespaceLister, podLister v1.PodLister) ([]*corev1.Pod, error) {
	pods := []*corev1.Pod{}
	ns, err := ListNamespacesInMesh(nsLister)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		podList, err := podLister.Pods(n.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve pods for namespace: %s error: %s", n.Name, err)
			return nil, err
		}
		for _, p := range podList {
			if SidecarInjected(p) == true {
				pods = append(pods, p)
			}
		}
	}
	return pods, nil
}

// ListServicesInMesh returns the list of Services in the mesh.
// Services in Namespaces returned by ListNamespacesInMesh are considered in the mesh.
func ListServicesInMesh(nsLister v1.NamespaceLister, svcLister v1.ServiceLister) ([]*corev1.Service, error) {
	services := []*corev1.Service{}
	ns, err := ListNamespacesInMesh(nsLister)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		serviceList, err := svcLister.Services(n.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve services for namespace: %s error: %s", n.Name, err)
			return nil, err
		}
		for _, s := range serviceList {
			if s.Name != "kubernetes" {
				services = append(services, s)
			}
		}
	}
	return services, nil
}

func IsEndpointInMesh(ea *corev1.EndpointAddress, podLister v1.PodLister) bool {
	if ea != nil && ea.TargetRef != nil {
		if ea.TargetRef.Kind == "Pod" {
			podList, err := podLister.Pods(ea.TargetRef.Namespace).List(labels.Everything())
			if err != nil {
				glog.Errorf("Failed to retrieve pods for namespace: %s error: %s", ea.TargetRef.Namespace, err)
				return false
			}
			for _, p := range podList {
				if p.Name == ea.TargetRef.Name && SidecarInjected(p) == true {
					return true
				}
			}
		}
	}
	return false
}

// ListEndpointsInMesh returns the list of Endpoints in the mesh.
// Endpoints in Namespaces returned by ListNamespacesInMesh are considered in the mesh.
func ListEndpointsInMesh(nsLister v1.NamespaceLister, epLister v1.EndpointsLister) ([]*corev1.Endpoints, error) {
	endpoints := []*corev1.Endpoints{}
	ns, err := ListNamespacesInMesh(nsLister)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		endpointList, err := epLister.Endpoints(n.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve endpoints for namespace: %s error: %s", n.Name, err)
			return nil, err
		}
		for _, s := range endpointList {
			if s.Name != kubernetesServiceName {
				endpoints = append(endpoints, s)
			}
		}
	}
	return endpoints, nil
}

// ComputeID returns MD5 checksum of the Note struct which can be used as
// ID for the note.
func ComputeID(n *apiv1.Note) string {
	return fmt.Sprintf("%x", structhash.Md5(n, 1))
}

// ListVirtualServices returns a list of VirtualService resources in the mesh.
func ListVirtualServicesInMesh(nsLister v1.NamespaceLister,
	vsLister netv1alpha3.VirtualServiceLister) ([]*v1alpha3.VirtualService, error) {
	virtualServices := []*v1alpha3.VirtualService{}
	ns, err := ListNamespacesInMesh(nsLister)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		virtServiceList, err := vsLister.VirtualServices(n.Name).List(labels.Everything())
		if err != nil {
			glog.Errorf("Failed to retrieve VirtualServices for namespace: %s error: %s", n.Name, err)
			return nil, err
		}
		virtualServices = append(virtualServices, virtServiceList...)
	}
	return virtualServices, nil
}

// ConvertHostnameToFQDN returns the FQDN if a short name is passed
func ConvertHostnameToFQDN(hostname string, namespace string) (string, error) {
	if (hostname == "") || (namespace == "") {
		err := errors.New("hostname and namespace cannot be empty")
		return "", err
	}
	if strings.HasPrefix(hostname, "*") {
		return hostname, nil
	}
	if strings.Contains(hostname, ".") {
		return hostname, nil
	}
	// need to return Fully Qualified Domain Name
	return hostname + "." + namespace + KubernetesDomainSuffix, nil
}

// ProxyStatusPort extracts status port from the cmd arguments for a given container,
// as per Istio 1.1 doc, global.proxy.statusPort https://istio.io/docs/reference/config/installation-options-changes/
func ProxyStatusPort(container corev1.Container) (uint32, error) {
	var statusPort uint32 = kubernetesProxyStatusPortDefault
	args := container.Args
	for index, key := range args {
		// Key we are looking for - hopefully followed by an argument specifying its value. If not, return default
		if key == kubernetesProxyStatusPort && index < len(args)-1 {
			// Next entry should be the port...
			overridePort, err := strconv.ParseUint(args[index+1], 10, 32)
			if err != nil {
				return statusPort, err
			}
			return uint32(overridePort), nil
		}
	}
	return statusPort, errors.New("cannot find proxy status port.")
}

package util

// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"net"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/types"
	meshconfig "istio.io/api/mesh/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// per-sidecar policy and status
var (
	alwaysValidFunc = func(value string) error {
		return nil
	}

	annotationRegistry = []*registeredAnnotation{
		{"sidecar.istio.io/inject", alwaysValidFunc},
		{"sidecar.istio.io/status", alwaysValidFunc},
		{"sidecar.istio.io/proxyImage", alwaysValidFunc},
		{"sidecar.istio.io/interceptionMode", validateInterceptionMode},
		{"status.sidecar.istio.io/port", validateStatusPort},
		{"readiness.status.sidecar.istio.io/initialDelaySeconds", validateUInt32},
		{"readiness.status.sidecar.istio.io/periodSeconds", validateUInt32},
		{"readiness.status.sidecar.istio.io/failureThreshold", validateUInt32},
		{"readiness.status.sidecar.istio.io/applicationPorts", validateReadinessApplicationPorts},
		{"traffic.sidecar.istio.io/includeOutboundIPRanges", ValidateIncludeIPRanges},
		{"traffic.sidecar.istio.io/excludeOutboundIPRanges", ValidateExcludeIPRanges},
		{"traffic.sidecar.istio.io/includeInboundPorts", ValidateIncludeInboundPorts},
		{"traffic.sidecar.istio.io/excludeInboundPorts", ValidateExcludeInboundPorts},
		{"traffic.sidecar.istio.io/kubevirtInterfaces", alwaysValidFunc},
	}

	annotationPolicy = annotationRegistry[0]
	annotationStatus = annotationRegistry[1]
)

type annotationValidationFunc func(value string) error

func validateAnnotations(annotations map[string]string) (err error) {
	for _, validator := range annotationRegistry {
		if e := validator.validate(annotations); e != nil {
			return e
		}
	}
	return
}

type registeredAnnotation struct {
	name      string
	validator annotationValidationFunc
}

func (v *registeredAnnotation) getValueOrDefault(annotations map[string]string, defaultValue string) string {
	if val, ok := annotations[v.name]; ok {
		return val
	}
	return defaultValue
}

func (v *registeredAnnotation) validate(annotations map[string]string) error {
	if val, ok := annotations[v.name]; ok {
		if err := v.validator(val); err != nil {
			return fmt.Errorf("injection failed. Invalid value for annotation %s: %s. Error: %v", v.name, val, err)
		}
	}
	return nil
}

// InjectionPolicy determines the policy for injecting the
// sidecar proxy into the watched namespace(s).
type InjectionPolicy string

const (
	// InjectionPolicyDisabled specifies that the sidecar injector
	// will not inject the sidecar into resources by default for the
	// namespace(s) being watched. Resources can enable injection
	// using the "sidecar.istio.io/inject" annotation with value of
	// true.
	InjectionPolicyDisabled InjectionPolicy = "disabled"

	// InjectionPolicyEnabled specifies that the sidecar injector will
	// inject the sidecar into resources by default for the
	// namespace(s) being watched. Resources can disable injection
	// using the "sidecar.istio.io/inject" annotation with value of
	// false.
	InjectionPolicyEnabled InjectionPolicy = "enabled"
)

// Defaults values for injecting istio proxy into kubernetes
// resources.
const (
	DefaultSidecarProxyUID              = uint64(1337)
	DefaultVerbosity                    = 2
	DefaultImagePullPolicy              = "IfNotPresent"
	DefaultStatusPort                   = 15020
	DefaultReadinessInitialDelaySeconds = 1
	DefaultReadinessPeriodSeconds       = 2
	DefaultReadinessFailureThreshold    = 30
	DefaultIncludeIPRanges              = "*"
	DefaultIncludeInboundPorts          = "*"
	DefaultkubevirtInterfaces           = ""
)

const (
	// ProxyContainerName is used by e2e integration tests for fetching logs
	ProxyContainerName = "istio-proxy"
)

// Aspenmesh inserts:
//---------------------------------------------
// Copied from other isto file locations:
const (
	sidecarTemplateDelimBegin = "[["
	sidecarTemplateDelimEnd   = "]]"
	InterceptionNone string = "NONE"
)

// End Aspenmesh inserts
//---------------------------------------------

// SidecarInjectionSpec collects all container types and volumes for
// sidecar mesh injection
type SidecarInjectionSpec struct {
	// RewriteHTTPProbe indicates whether Kubernetes HTTP prober in the PodSpec
	// will be rewritten to be redirected by pilot agent.
	RewriteAppHTTPProbe bool                          `yaml:"rewriteAppHTTPProbe"`
	InitContainers      []corev1.Container            `yaml:"initContainers"`
	Containers          []corev1.Container            `yaml:"containers"`
	Volumes             []corev1.Volume               `yaml:"volumes"`
	DNSConfig           *corev1.PodDNSConfig          `yaml:"dnsConfig"`
	ImagePullSecrets    []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
}

// SidecarTemplateData is the data object to which the templated
// version of `SidecarInjectionSpec` is applied.
type SidecarTemplateData struct {
	DeploymentMeta *metav1.ObjectMeta
	ObjectMeta     *metav1.ObjectMeta
	Spec           *corev1.PodSpec
	ProxyConfig    *meshconfig.ProxyConfig
	MeshConfig     *meshconfig.MeshConfig
}

// InitImageName returns the fully qualified image name for the istio
// init image given a docker hub and tag and debug flag
func InitImageName(hub string, tag string, _ bool) string {
	return hub + "/proxy_init:" + tag
}

// ProxyImageName returns the fully qualified image name for the istio
// proxy image given a docker hub and tag and whether to use debug or not.
func ProxyImageName(hub string, tag string, debug bool) string {
	// Allow overriding the proxy image.
	if debug {
		return hub + "/proxy_debug:" + tag
	}
	return hub + "/proxyv2:" + tag
}

// Params describes configurable parameters for injecting istio proxy
// into a kubernetes resource.
type Params struct {
	InitImage                    string                 `json:"initImage"`
	RewriteAppHTTPProbe          bool                   `json:"rewriteAppHTTPProbe"`
	ProxyImage                   string                 `json:"proxyImage"`
	Verbosity                    int                    `json:"verbosity"`
	SidecarProxyUID              uint64                 `json:"sidecarProxyUID"`
	Version                      string                 `json:"version"`
	EnableCoreDump               bool                   `json:"enableCoreDump"`
	DebugMode                    bool                   `json:"debugMode"`
	Privileged                   bool                   `json:"privileged"`
	Mesh                         *meshconfig.MeshConfig `json:"-"`
	ImagePullPolicy              string                 `json:"imagePullPolicy"`
	StatusPort                   int                    `json:"statusPort"`
	ReadinessInitialDelaySeconds uint32                 `json:"readinessInitialDelaySeconds"`
	ReadinessPeriodSeconds       uint32                 `json:"readinessPeriodSeconds"`
	ReadinessFailureThreshold    uint32                 `json:"readinessFailureThreshold"`
	SDSEnabled                   bool                   `json:"sdsEnabled"`
	EnableSdsTokenMount          bool                   `json:"enableSdsTokenMount"`
	// Comma separated list of IP ranges in CIDR form. If set, only redirect outbound traffic to Envoy for these IP
	// ranges. All outbound traffic can be redirected with the wildcard character "*". Defaults to "*".
	IncludeIPRanges string `json:"includeIPRanges"`
	// Comma separated list of IP ranges in CIDR form. If set, outbound traffic will not be redirected for
	// these IP ranges. Exclusions are only applied if configured to redirect all outbound traffic. By default,
	// no IP ranges are excluded.
	ExcludeIPRanges string `json:"excludeIPRanges"`
	// Comma separated list of inbound ports for which traffic is to be redirected to Envoy. All ports can be
	// redirected with the wildcard character "*". Defaults to "*".
	IncludeInboundPorts string `json:"includeInboundPorts"`
	// Comma separated list of inbound ports. If set, inbound traffic will not be redirected for those ports.
	// Exclusions are only applied if configured to redirect all inbound traffic. By default, no ports are excluded.
	ExcludeInboundPorts string `json:"excludeInboundPorts"`
	// Comma separated list of virtual interfaces whose inbound traffic (from VM) will be treated as outbound
	// By default, no interfaces are configured.
	KubevirtInterfaces string `json:"kubevirtInterfaces"`
}

// Validate validates the parameters and returns an error if there is configuration issue.
func (p *Params) Validate() error {
	if err := ValidateIncludeIPRanges(p.IncludeIPRanges); err != nil {
		return err
	}
	if err := ValidateExcludeIPRanges(p.ExcludeIPRanges); err != nil {
		return err
	}
	if err := ValidateIncludeInboundPorts(p.IncludeInboundPorts); err != nil {
		return err
	}
	return ValidateExcludeInboundPorts(p.ExcludeInboundPorts)
}

// Config specifies the sidecar injection configuration This includes
// the sidecar template and cluster-side injection policy. It is used
// by kube-inject, sidecar injector, and http endpoint.
type Config struct {
	Policy InjectionPolicy `json:"policy"`

	// Template is the templated version of `SidecarInjectionSpec` prior to
	// expansion over the `SidecarTemplateData`.
	Template string `json:"template"`

	// NeverInjectSelector: Refuses the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	// Takes precedence over AlwaysInjectSelector.
	NeverInjectSelector []metav1.LabelSelector `json:"neverInjectSelector"`

	// AlwaysInjectSelector: Forces the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	AlwaysInjectSelector []metav1.LabelSelector `json:"alwaysInjectSelector"`
}

func validateCIDRList(cidrs string) error {
	if len(cidrs) > 0 {
		for _, cidr := range strings.Split(cidrs, ",") {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("failed parsing cidr '%s': %v", cidr, err)
			}
		}
	}
	return nil
}

func splitPorts(portsString string) []string {
	return strings.Split(portsString, ",")
}

func parsePort(portStr string) (int, error) {
	port, err := strconv.ParseUint(strings.TrimSpace(portStr), 10, 16)
	if err != nil {
		return 0, fmt.Errorf("failed parsing port '%d': %v", port, err)
	}
	return int(port), nil
}

func parsePorts(portsString string) ([]int, error) {
	portsString = strings.TrimSpace(portsString)
	ports := make([]int, 0)
	if len(portsString) > 0 {
		for _, portStr := range splitPorts(portsString) {
			port, err := parsePort(portStr)
			if err != nil {
				return nil, fmt.Errorf("failed parsing port '%d': %v", port, err)
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}

func validatePortList(parameterName, ports string) error {
	if _, err := parsePorts(ports); err != nil {
		return fmt.Errorf("%s invalid: %v", parameterName, err)
	}
	return nil
}

// validateInterceptionMode validates the interceptionMode annotation
func validateInterceptionMode(mode string) error {

	// Aspenmesh modification
	switch mode {
	case meshconfig.ProxyConfig_REDIRECT.String():
	case meshconfig.ProxyConfig_TPROXY.String():
	case string(InterceptionNone): // not a global mesh config - must be enabled for each sidecar
	default:
		return fmt.Errorf("interceptionMode invalid, use REDIRECT,TPROXY,NONE: %v", mode)
	}
	return nil
}


// ValidateIncludeIPRanges validates the includeIPRanges parameter
func ValidateIncludeIPRanges(ipRanges string) error {
	if ipRanges != "*" {
		if e := validateCIDRList(ipRanges); e != nil {
			return fmt.Errorf("includeIPRanges invalid: %v", e)
		}
	}
	return nil
}

// ValidateExcludeIPRanges validates the excludeIPRanges parameter
func ValidateExcludeIPRanges(ipRanges string) error {
	if e := validateCIDRList(ipRanges); e != nil {
		return fmt.Errorf("excludeIPRanges invalid: %v", e)
	}
	return nil
}

func validateReadinessApplicationPorts(ports string) error {
	if ports != "*" {
		return validatePortList("readinessApplicationPorts", ports)
	}
	return nil
}

// ValidateIncludeInboundPorts validates the includeInboundPorts parameter
func ValidateIncludeInboundPorts(ports string) error {
	if ports != "*" {
		return validatePortList("includeInboundPorts", ports)
	}
	return nil
}

// ValidateExcludeInboundPorts validates the excludeInboundPorts parameter
func ValidateExcludeInboundPorts(ports string) error {
	return validatePortList("excludeInboundPorts", ports)
}

// validateStatusPort validates the statusPort parameter
func validateStatusPort(port string) error {
	if _, e := parsePort(port); e != nil {
		return fmt.Errorf("excludeInboundPorts invalid: %v", e)
	}
	return nil
}

// validateUInt32 validates that the given annotation value is a positive integer.
func validateUInt32(value string) error {
	_, err := strconv.ParseUint(value, 10, 32)
	return err
}

func formatDuration(in *types.Duration) string {
	dur, err := types.DurationFromProto(in)
	if err != nil {
		return "1s"
	}
	return dur.String()
}

func isset(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func directory(filepath string) string {
	dir, _ := path.Split(filepath)
	return dir
}

func injectionData(sidecarTemplate, version string, deploymentMetadata *metav1.ObjectMeta, spec *corev1.PodSpec,
	metadata *metav1.ObjectMeta, proxyConfig *meshconfig.ProxyConfig, meshConfig *meshconfig.MeshConfig) (
	*SidecarInjectionSpec, string, error) { // nolint: lll
	if err := validateAnnotations(metadata.GetAnnotations()); err != nil {
		glog.Infof("Invalid annotations: %v %v\n", err, metadata.GetAnnotations())
		return nil, "", err
	}

	data := SidecarTemplateData{
		DeploymentMeta: deploymentMetadata,
		ObjectMeta:     metadata,
		Spec:           spec,
		ProxyConfig:    proxyConfig,
		MeshConfig:     meshConfig,
	}

	funcMap := template.FuncMap{
		"formatDuration":      formatDuration,
		"isset":               isset,
		"excludeInboundPort":  excludeInboundPort,
		"includeInboundPorts": includeInboundPorts,
		"kubevirtInterfaces":  kubevirtInterfaces,
		"applicationPorts":    applicationPorts,
		"annotation":          annotation,
		"valueOrDefault":      valueOrDefault,
		"toJSON":              toJSON,
		"toJson":              toJSON, // Used by, e.g. Istio 1.0.5 template sidecar-injector-configmap.yaml
		"fromJSON":            fromJSON,
		"toYaml":              toYaml,
		"indent":              indent,
		"directory":           directory,
	}

	var tmpl bytes.Buffer
	temp := template.New("inject").Delims(sidecarTemplateDelimBegin, sidecarTemplateDelimEnd)
	t, err := temp.Funcs(funcMap).Parse(sidecarTemplate)
	if err != nil {
		glog.Infof("Failed to parse template: %v %v\n", err, sidecarTemplate)
		return nil, "", err
	}
	if err := t.Execute(&tmpl, &data); err != nil {
		glog.Infof("Invalid template: %v %v\n", err, sidecarTemplate)
		return nil, "", err
	}

	var sic SidecarInjectionSpec
	if err := yaml.Unmarshal(tmpl.Bytes(), &sic); err != nil {
		glog.Warningf("Failed to unmarshall template %v %s", err, tmpl.String())
		return nil, "", err
	}

	// set sidecar --concurrency
	// ASPENMESH - Comment out as not needed
	// applyConcurrency(sic.Containers)

	status := &SidecarInjectionStatus{Version: version}
	for _, c := range sic.InitContainers {
		status.InitContainers = append(status.InitContainers, c.Name)
	}
	for _, c := range sic.Containers {
		status.Containers = append(status.Containers, c.Name)
	}
	for _, c := range sic.Volumes {
		status.Volumes = append(status.Volumes, c.Name)
	}
	for _, c := range sic.ImagePullSecrets {
		status.ImagePullSecrets = append(status.ImagePullSecrets, c.Name)
	}
	statusAnnotationValue, err := json.Marshal(status)
	if err != nil {
		return nil, "", fmt.Errorf("error encoded injection status: %v", err)
	}
	return &sic, string(statusAnnotationValue), nil
}

func getPortsForContainer(container corev1.Container) []string {
	parts := make([]string, 0)
	for _, p := range container.Ports {
		parts = append(parts, strconv.Itoa(int(p.ContainerPort)))
	}
	return parts
}

func getContainerPorts(containers []corev1.Container, shouldIncludePorts func(corev1.Container) bool) string {
	parts := make([]string, 0)
	for _, c := range containers {
		if shouldIncludePorts(c) {
			parts = append(parts, getPortsForContainer(c)...)
		}
	}

	return strings.Join(parts, ",")
}

func applicationPorts(containers []corev1.Container) string {
	return getContainerPorts(containers, func(c corev1.Container) bool {
		return c.Name != ProxyContainerName
	})
}

func includeInboundPorts(containers []corev1.Container) string {
	// Include the ports from all containers in the deployment.
	return getContainerPorts(containers, func(corev1.Container) bool { return true })
}

func kubevirtInterfaces(s string) string {
	return s
}

func toJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}

	ba, err := json.Marshal(m)
	if err != nil {
		glog.Warningf("Unable to marshal %v", m)
		return "{}"
	}

	return string(ba)
}

func fromJSON(j string) interface{} {
	var m interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		glog.Warningf("Unable to unmarshal %s", j)
		return "{}"
	}

	glog.Warningf("%v", m)
	return m
}

func indent(spaces int, source string) string {
	res := strings.Split(source, "\n")
	for i, line := range res {
		if i > 0 {
			res[i] = fmt.Sprintf(fmt.Sprintf("%% %ds%%s", spaces), "", line)
		}
	}
	return strings.Join(res, "\n")
}

func toYaml(value interface{}) string {
	y, err := yaml.Marshal(value)
	if err != nil {
		glog.Warningf("Unable to marshal %v", value)
		return ""
	}

	return string(y)
}

func annotation(meta metav1.ObjectMeta, name string, defaultValue interface{}) string {
	value, ok := meta.Annotations[name]
	if !ok {
		value = fmt.Sprint(defaultValue)
	}
	return value
}

func excludeInboundPort(port interface{}, excludedInboundPorts string) string {
	portStr := strings.TrimSpace(fmt.Sprint(port))
	if len(portStr) == 0 || portStr == "0" {
		// Nothing to do.
		return excludedInboundPorts
	}

	// Exclude the readiness port if not already excluded.
	ports := splitPorts(excludedInboundPorts)
	outPorts := make([]string, 0, len(ports))
	for _, port := range ports {
		if port == portStr {
			// The port is already excluded.
			return excludedInboundPorts
		}
		port = strings.TrimSpace(port)
		if len(port) > 0 {
			outPorts = append(outPorts, port)
		}
	}

	// The port was not already excluded - exclude it now.
	outPorts = append(outPorts, portStr)
	return strings.Join(outPorts, ",")
}

func valueOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// SidecarInjectionStatus contains basic information about the
// injected sidecar. This includes the names of added containers and
// volumes.
type SidecarInjectionStatus struct {
	Version          string   `json:"version"`
	InitContainers   []string `json:"initContainers"`
	Containers       []string `json:"containers"`
	Volumes          []string `json:"volumes"`
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// helper function to generate a template version identifier from a
// hash of the un-executed template contents.
func sidecarTemplateVersionHash(in string) string {
	hash := sha256.Sum256([]byte(in))
	return hex.EncodeToString(hash[:])
}

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}

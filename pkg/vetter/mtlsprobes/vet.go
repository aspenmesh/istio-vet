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

// Package mtlsprobes inspects if Istio mTLS and liveness probes are enabled for
// any Pods in the mesh.
package mtlsprobes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"
	proxyconfig "istio.io/api/proxy/v1/config"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

const (
	vetterID                 = "MtlsProbes"
	mtlsProbesNoteType       = "mtls-probes-incompatible"
	mtlsLivenessProbeSummary = "mTLS and liveness probe incompatible - ${pod_name}"
	mtlsLivenessProbeMsg     = "The pod ${pod_name} in namespace ${namespace} uses" +
		" liveness probe which is incompatible with mTLS. Consider disabling the" +
		" liveness probe or mTLS."
	mtlsReadinessProbeSummary = "mTLS and readiness probe incompatible - ${pod_name}"
	mtlsReadinessProbeMsg     = "The pod ${pod_name} in namespace ${namespace} uses" +
		" readiness probe which is incompatible with mTLS. Consider disabling the" +
		" readiness probe or mTLS."
	mtlsDisabledSummary = "mTLS is disabled. Enable it to use \"" +
		vetterID + "\" vetter"
)

// MtlsProbes implements Vetter interface
type MtlsProbes struct {
	info apiv1.Info
}

func mtlsEnabled(c string) bool {
	var cfg proxyconfig.MeshConfig
	if err := util.ApplyYAML(c, &cfg); err != nil {
		glog.Errorf("Failed to parse yaml mesh config: %s", err)
		return false
	}
	return cfg.GetAuthPolicy() != 0
}

// Vet returns the list of generated notes
func (m *MtlsProbes) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	cm, err :=
		c.CoreV1().ConfigMaps(util.IstioNamespace).Get(util.IstioConfigMap,
			metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Failed to retrieve configmap: %s error: %s", util.IstioConfigMap, err)
		return nil, err
	}
	config := cm.Data[util.IstioConfigMapKey]
	if len(config) == 0 {
		return nil, nil
	}
	if mtlsEnabled(config) == false {
		notes = append(notes, &apiv1.Note{
			Type:    mtlsProbesNoteType,
			Summary: mtlsDisabledSummary,
			Level:   apiv1.NoteLevel_INFO})
		return notes, nil
	}
	pods, err := util.ListPodsInMesh(c)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			mtlsProbesNoteType); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	for _, p := range pods {
		if util.SidecarInjected(p) == true {
			cList := p.Spec.Containers
			for _, c := range cList {
				if c.LivenessProbe != nil && c.LivenessProbe.Exec == nil {
					notes = append(notes, &apiv1.Note{
						Type:    mtlsProbesNoteType,
						Summary: mtlsLivenessProbeSummary,
						Msg:     mtlsLivenessProbeMsg,
						Level:   apiv1.NoteLevel_ERROR,
						Attr: map[string]string{
							"pod_name":  p.Name,
							"namespace": p.Namespace}})
				} else if c.ReadinessProbe != nil && c.ReadinessProbe.Exec == nil {
					notes = append(notes, &apiv1.Note{
						Type:    mtlsProbesNoteType,
						Summary: mtlsReadinessProbeSummary,
						Msg:     mtlsReadinessProbeMsg,
						Level:   apiv1.NoteLevel_ERROR,
						Attr: map[string]string{
							"pod_name":  p.Name,
							"namespace": p.Namespace}})
				}
			}
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

// Info returns information about the vetter
func (m *MtlsProbes) Info() *apiv1.Info {
	return &m.info
}

// NewVetter returns "MtlsProbes" which implements Vetter Interface
func NewVetter() *MtlsProbes {
	return &MtlsProbes{info: apiv1.Info{Id: vetterID, Version: "0.1.0"}}
}

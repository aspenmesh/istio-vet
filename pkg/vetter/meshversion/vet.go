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

package meshversion

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

const (
	latest_tag                         = "latest"
	istio_component_mismatch_note_type = "istio-component-mismatch"
	istio_component_mismatch_summary   = "Mismatched istio component versions - ${component_name}"
	istio_component_mismatch_msg       = "Istio component ${component_name} is running version ${component_version}" +
		" but your environment is running istio version ${istio_version}." +
		" Consider upgrading the component ${component_name} "
	sidecar_mismatch_note_type = "sidecar-version-mismatch"
	sidecar_mismatch_summary   = "Mismatched sidecar version - ${pod_name}"
	sidecar_mismatch_msg       = "The pod ${pod_name} in namespace ${namespace}" +
		" is running with sidecar proxy version ${sidecar_version}" +
		" but your environment is running istio version" +
		" ${istio_version}. Consider upgrading the sidecar proxy in the pod."
	missing_version_note_type    = "missing-version"
	missing_version_note_summary = "Missing version information"
	missing_version_note_msg     = "Cannot determine mesh version"
)

type meshVersion struct {
	info apiv1.Info
}

func getImageVersion(c kubernetes.Interface, namespace, deployment, container string) (string, error) {
	opts := metav1.GetOptions{}
	d, err := c.ExtensionsV1beta1().Deployments(namespace).Get(deployment, opts)
	if err != nil {
		glog.Errorf("Failed to retrieve deployment: %s in namespace: %s error: %s",
			deployment, namespace, err)
		return "", err
	}
	return util.ImageTag(container, d.Spec.Template.Spec)
}

func istioVersion(c kubernetes.Interface) (string, error) {
	return getImageVersion(c, util.IstioNamespace, util.IstioMixerDeploymentName,
		util.IstioMixerContainerName)
}

func (m *meshVersion) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	ver, err := istioVersion(c)
	if err != nil {
		return nil, err
	}
	if ver == latest_tag {
		notes = append(notes, &apiv1.Note{
			Type:    missing_version_note_type,
			Summary: missing_version_note_summary,
			Msg:     missing_version_note_msg,
			Level:   apiv1.NoteLevel_INFO})

		return notes, nil
	}

	pilotVer, err := getImageVersion(c, util.IstioNamespace,
		util.IstioPilotDeploymentName, util.IstioPilotContainerName)
	if pilotVer != latest_tag && pilotVer != ver {
		notes = append(notes, &apiv1.Note{
			Type:    istio_component_mismatch_note_type,
			Summary: istio_component_mismatch_summary,
			Msg:     istio_component_mismatch_msg,
			Level:   apiv1.NoteLevel_WARNING,
			Attr: map[string]string{
				"component_name":    util.IstioPilotDeploymentName,
				"component_version": pilotVer,
				"istio_version":     ver}})
	}

	pods, err := util.ListPodsInMesh(c)
	if err != nil {
		return nil, err
	}
	for _, p := range pods {
		if util.SidecarInjected(p) == true {
			sideCarVer, err := util.ImageTag(util.IstioProxyContainerName, p.Spec)
			if err != nil || sideCarVer == latest_tag {
				continue
			}
			if sideCarVer != ver {
				notes = append(notes, &apiv1.Note{
					Type:    sidecar_mismatch_note_type,
					Summary: sidecar_mismatch_summary,
					Msg:     sidecar_mismatch_msg,
					Level:   apiv1.NoteLevel_WARNING,
					Attr: map[string]string{
						"pod_name":        p.Name,
						"namespace":       p.Namespace,
						"sidecar_version": sideCarVer,
						"istio_version":   ver}})
			}
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeId(notes[i])
	}

	return notes, nil
}

func (m *meshVersion) Info() *apiv1.Info {
	return &m.info
}

func NewVetter() *meshVersion {
	return &meshVersion{info: apiv1.Info{Id: "meshversion", Version: "0.1.0"}}
}

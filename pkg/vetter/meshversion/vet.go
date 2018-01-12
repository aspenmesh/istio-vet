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

// Package meshversion vets the versions of Istio components, Sidecar proxy
// and generates notes on version mismatch.
package meshversion

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/golang/glog"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

const (
	vetterID                       = "MeshVersion"
	latestTag                      = "latest"
	istioComponentMismatchNoteType = "istio-component-mismatch"
	istioComponentMismatchSummary  = "Mismatched istio component versions - ${component_name}"
	istioComponentMismatchMsg      = "Istio component ${component_name} is running version ${component_version}" +
		" but your environment is running istio version ${istio_version}." +
		" Consider upgrading the component ${component_name} "
	sidecarMismatchNoteType = "sidecar-version-mismatch"
	sidecarMismatchSummary  = "Mismatched sidecar version - ${pod_name}"
	sidecarMismatchMsg      = "The pod ${pod_name} in namespace ${namespace}" +
		" is running with sidecar proxy version ${sidecar_version}" +
		" but your environment is running istio version" +
		" ${istio_version}. Consider upgrading the sidecar proxy in the pod."
	missingVersionNoteType    = "missing-version"
	missingVersionNoteSummary = "Missing version information"
	missingVersionNoteMsg     = "Cannot determine mesh version"
)

// MeshVersion implements Vetter interface
type MeshVersion struct {
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

// Vet returns the list of generated notes
func (m *MeshVersion) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	ver, err := istioVersion(c)
	if err != nil {
		return nil, err
	}
	if ver == latestTag {
		notes = append(notes, &apiv1.Note{
			Type:    missingVersionNoteType,
			Summary: missingVersionNoteSummary,
			Msg:     missingVersionNoteMsg,
			Level:   apiv1.NoteLevel_INFO})

		return notes, nil
	}

	pilotVer, err := getImageVersion(c, util.IstioNamespace,
		util.IstioPilotDeploymentName, util.IstioPilotContainerName)
	if pilotVer != latestTag && pilotVer != ver {
		notes = append(notes, &apiv1.Note{
			Type:    istioComponentMismatchNoteType,
			Summary: istioComponentMismatchSummary,
			Msg:     istioComponentMismatchMsg,
			Level:   apiv1.NoteLevel_WARNING,
			Attr: map[string]string{
				"component_name":    util.IstioPilotDeploymentName,
				"component_version": pilotVer,
				"istio_version":     ver}})
	}

	pods, err := util.ListPodsInMesh(c)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			sidecarMismatchNoteType); n != nil {
			notes = append(notes, n)
			return notes, nil
		}
		return nil, err
	}
	for _, p := range pods {
		sideCarVer, err := util.ImageTag(util.IstioProxyContainerName, p.Spec)
		if err != nil || sideCarVer == latestTag {
			continue
		}
		if sideCarVer != ver {
			notes = append(notes, &apiv1.Note{
				Type:    sidecarMismatchNoteType,
				Summary: sidecarMismatchSummary,
				Msg:     sidecarMismatchMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"pod_name":        p.Name,
					"namespace":       p.Namespace,
					"sidecar_version": sideCarVer,
					"istio_version":   ver}})
		}
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

// Info returns information about the vetter
func (m *MeshVersion) Info() *apiv1.Info {
	return &m.info
}

// NewVetter returns "MeshVersion" which implements Vetter Interface
func NewVetter() *MeshVersion {
	return &MeshVersion{info: apiv1.Info{Id: vetterID, Version: "0.1.0"}}
}

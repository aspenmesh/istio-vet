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

// Package podsinmesh generates informational notes on system and user Pods in
// the mesh.
package podsinmesh

import (
	"strconv"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/aspenmesh/istio-vet/pkg/vetter/util"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
)

const (
	userPodCountNoteType   = "user-pod-count"
	userPodCountSummary    = "User pod count"
	userPodCountMsg        = "${user_pods_in_mesh} user pods in mesh out of ${num_user_pods}"
	systemPodCountNoteType = "system-pod-count"
	systemPodCountSummary  = "System pod count"
	systemPodCountMsg      = "${num_system_pods} system pods out of mesh"
)

// MeshStats implements Vetter interface
type MeshStats struct {
	info apiv1.Info
}

// Vet returns the list of generated notes
func (m *MeshStats) Vet(c kubernetes.Interface) ([]*apiv1.Note, error) {
	opts := metav1.ListOptions{}
	ns, err := c.CoreV1().Namespaces().List(opts)
	if err != nil {
		glog.Error("Failed to retrieve namespaces: ", err)
		return nil, err
	}
	var totalUserPods, totalUserPodsInMesh, totalSystemPods int
	for _, n := range ns.Items {
		podList, err := c.CoreV1().Pods(n.Name).List(opts)
		if err != nil {
			glog.Errorf("Failed to retrieve pods for namespace: %s : %s", n.Name, err)
			return nil, err
		}
		if util.ExemptedNamespace(n.Name) == false {
			totalUserPods += len(podList.Items)
			for _, p := range podList.Items {
				if util.SidecarInjected(p) == true {
					totalUserPodsInMesh++
				}
			}
		} else {
			totalSystemPods += len(podList.Items)
		}
	}

	notes := []*apiv1.Note{
		&apiv1.Note{
			Type:    userPodCountNoteType,
			Summary: userPodCountSummary,
			Msg:     userPodCountMsg,
			Level:   apiv1.NoteLevel_INFO,
			Attr: map[string]string{
				"user_pods_in_mesh": strconv.Itoa(totalUserPodsInMesh),
				"num_user_pods":     strconv.Itoa(totalUserPods)}},
		&apiv1.Note{
			Type:    systemPodCountNoteType,
			Summary: systemPodCountSummary,
			Msg:     systemPodCountMsg,
			Level:   apiv1.NoteLevel_INFO,
			Attr: map[string]string{
				"num_system_pods": strconv.Itoa(totalSystemPods)}}}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

// Info returns information about the vetter
func (m *MeshStats) Info() *apiv1.Info {
	return &m.info
}

// NewVetter returns "meshStats" which implements Vetter Interface
func NewVetter() *MeshStats {
	return &MeshStats{info: apiv1.Info{Id: "podsinmesh", Version: "0.1.0"}}
}

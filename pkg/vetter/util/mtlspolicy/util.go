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
	"strings"

	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
)

// Service is the necesary components of a kubernetes service to look at auth
// policies and destination rules
type Service struct {
	Name      string
	Namespace string
}

// ServiceFromFqdn validates a kubernetes FQDN and returns a service with the
// name and namespace from a validated FQDN
func ServiceFromFqdn(fqdn string) (Service, error) {
	if !strings.HasSuffix(fqdn, util.KubernetesDomainSuffix) {
		return Service{}, errors.New("FQDN suffix unrecognized")
	}
	front := strings.TrimSuffix(fqdn, util.KubernetesDomainSuffix)
	parts := strings.Split(front, ".")
	if len(parts) != 2 || len(parts[0]) < 1 || len(parts[1]) < 1 {
		return Service{}, errors.New("FQDN does not have name and namespace")
	}
	return Service{Name: parts[0], Namespace: parts[1]}, nil
}

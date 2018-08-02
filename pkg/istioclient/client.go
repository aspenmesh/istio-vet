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

// Package istioclient provides utilities for creating a istio client.
package istioclient

import (
	"github.com/aspenmesh/istio-client-go/pkg/client/clientset/versioned"
	"k8s.io/client-go/rest"
)

// New returns a new versioned Istio Client
func New(rc *rest.Config) (versioned.Interface, error) {
	c, err := versioned.NewForConfig(rc)
	if err != nil {
		return nil, err
	}
	return c, nil
}

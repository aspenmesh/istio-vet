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

// Package meshclient provides utilities for creating a kubernetes client.
package meshclient

import (
	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeConfig clientcmd.ClientConfig

func New() (kubernetes.Interface, error) {
	cs, err := GetClientset()
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func BindKubeConfigToFlags(flags *pflag.FlagSet) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	configOverrides := clientcmd.ConfigOverrides{}

	clientcmd.BindOverrideFlags(&configOverrides, flags, clientcmd.RecommendedConfigOverrideFlags("k8s-"))

	kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &configOverrides)
}

func GetClientset() (*kubernetes.Clientset, error) {

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

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

package cmd

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	istioinformer "istio.io/client-go/pkg/informers/externalversions"
	"k8s.io/client-go/informers"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/aspenmesh/istio-vet/pkg/istioclient"
	"github.com/aspenmesh/istio-vet/pkg/meshclient"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/applabel"
	"github.com/aspenmesh/istio-vet/pkg/vetter/conflictingvirtualservicehost"
	"github.com/aspenmesh/istio-vet/pkg/vetter/danglingroutedestinationhost"
	"github.com/aspenmesh/istio-vet/pkg/vetter/meshversion"
	"github.com/aspenmesh/istio-vet/pkg/vetter/mtlsprobes"
	"github.com/aspenmesh/istio-vet/pkg/vetter/podsinmesh"
	"github.com/aspenmesh/istio-vet/pkg/vetter/serviceassociation"
	"github.com/aspenmesh/istio-vet/pkg/vetter/serviceportprefix"
)

func printNote(level, summary, msg string) {
	if len(summary) > 0 {
		fmt.Printf("%s\n", summary)
		if len(msg) > 0 {
			b := make([]byte, len(summary))
			for i := range b {
				b[i] = '='
			}
			fmt.Printf("%s\n", b)
		} else {
			fmt.Println()
		}
	}
	if len(msg) > 0 {
		fmt.Printf("%s: %s\n\n", level, msg)
	}
}

type metaInformerFactory struct {
	k8s   informers.SharedInformerFactory
	istio istioinformer.SharedInformerFactory
}

func (m *metaInformerFactory) K8s() informers.SharedInformerFactory {
	return m.k8s
}
func (m *metaInformerFactory) Istio() istioinformer.SharedInformerFactory {
	return m.istio
}

func vet(cmd *cobra.Command, args []string) error {
	k8sClient, err := meshclient.New()
	if err != nil {
		return err
	}
	istioClient, err := istioclient.New(k8sClient.Config())
	if err != nil {
		return err
	}

	kubeInformerFactory := informers.NewSharedInformerFactory(k8sClient, 0)
	istioInformerFactory := istioinformer.NewSharedInformerFactory(istioClient, 0)
	informerFactory := &metaInformerFactory{
		k8s:   kubeInformerFactory,
		istio: istioInformerFactory,
	}

	vList := []vetter.Vetter{
		vetter.Vetter(podsinmesh.NewVetter(informerFactory)),
		vetter.Vetter(meshversion.NewVetter(informerFactory)),
		vetter.Vetter(mtlsprobes.NewVetter(informerFactory)),
		vetter.Vetter(applabel.NewVetter(informerFactory)),
		vetter.Vetter(serviceportprefix.NewVetter(informerFactory)),
		vetter.Vetter(serviceassociation.NewVetter(informerFactory)),
		vetter.Vetter(danglingroutedestinationhost.NewVetter(informerFactory)),
		vetter.Vetter(conflictingvirtualservicehost.NewVetter(informerFactory)),
		// obsolete in Istio 1.5
		// vetter.Vetter(invalidserviceforjwtpolicy.NewVetter(informerFactory)),
	}

	stopCh := make(chan struct{})

	kubeInformerFactory.Start(stopCh)
	oks := kubeInformerFactory.WaitForCacheSync(stopCh)
	for inf, ok := range oks {
		if !ok {
			glog.Fatalf("Failed to sync %s", inf)
		}
	}

	istioInformerFactory.Start(stopCh)
	oks = istioInformerFactory.WaitForCacheSync(stopCh)
	for inf, ok := range oks {
		if !ok {
			glog.Fatalf("Failed to sync %s", inf)
		}
	}
	// Just run through once
	close(stopCh)

	for _, v := range vList {
		nList, err := v.Vet()
		if err != nil {
			fmt.Printf("Vetter: \"%s\" reported error: %s\n", v.Info().GetId(), err)
			continue
		}
		if len(nList) > 0 {
			for i := range nList {
				var ts []string
				for k, v := range nList[i].Attr {
					ts = append(ts, "${"+k+"}", v)
				}
				r := strings.NewReplacer(ts...)
				summary := r.Replace(nList[i].GetSummary())
				msg := r.Replace(nList[i].GetMsg())
				printNote(nList[i].GetLevel().String(), summary, msg)
			}
		} else {
			fmt.Printf("Vetter \"%s\" ran successfully and generated no notes\n\n", v.Info().GetId())
		}
	}

	return nil
}

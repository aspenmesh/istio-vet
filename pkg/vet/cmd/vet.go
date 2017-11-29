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

	"github.com/spf13/cobra"

	"github.com/aspenmesh/istio-vet/pkg/meshclient"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/meshversion"
	"github.com/aspenmesh/istio-vet/pkg/vetter/mtlsprobes"
	"github.com/aspenmesh/istio-vet/pkg/vetter/podsinmesh"
	"github.com/aspenmesh/istio-vet/pkg/vetter/serviceportprefix"
)

func vet(cmd *cobra.Command, args []string) error {
	cli, err := meshclient.New()
	if err != nil {
		return err
	}
	vList := []vetter.Vetter{
		vetter.Vetter(podsinmesh.NewVetter()),
		vetter.Vetter(meshversion.NewVetter()),
		vetter.Vetter(mtlsprobes.NewVetter()),
		vetter.Vetter(serviceportprefix.NewVetter())}

	for _, v := range vList {
		nList, err := v.Vet(cli)
		if err != nil {
			fmt.Printf("Vetter:  %+v reported error: %s", v.Info(), err)
			continue
		}
		for i := range nList {
			var ts []string
			for k, v := range nList[i].Attr {
				ts = append(ts, "${"+k+"}", v)
			}
			r := strings.NewReplacer(ts...)
			summary := r.Replace(nList[i].GetSummary())
			msg := r.Replace(nList[i].GetMsg())
			fmt.Printf("%s\n", summary)
			fmt.Printf("==========================\n")
			fmt.Printf("%s: %s\n\n", nList[i].GetLevel().String(), msg)
		}
	}

	return nil
}

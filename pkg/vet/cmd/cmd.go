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
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aspenmesh/istio-vet/pkg/meshclient"
	"github.com/aspenmesh/istio-vet/pkg/util/logs"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	// DefaultConfigFile is the default config file for vet tool
	DefaultConfigFile = "/etc/istio/vet.yaml"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "vet",
	Short: "Runs vet command",
	Long: `Runs vet command.

Vet is a diagnostic tool for validating the configuration of Istio
and applications deployed in the mesh.

For more details, see 'https://github.com/aspenmesh/istio-vet'
`,
	RunE: vet,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Load flags added by other packages
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Normalize those other flags. (glog likes to use _)
	pflag.CommandLine.SetNormalizeFunc(externalFlagNormalize)
	// Copy those flags into root command
	meshclient.BindKubeConfigToFlags(RootCmd.PersistentFlags())
	RootCmd.PersistentFlags().AddFlagSet(pflag.CommandLine)
}

// WordSepNormalizeFunc changes all flags that contain "_" separators
func externalFlagNormalize(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	logs.InitLogs()
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("vet_config")          // name of config file (without extension)
	viper.AddConfigPath(".")                   // adding local directory as first search path
	viper.AddConfigPath("$HOME/.config/istio") // adding home directory as first search path
	viper.SetEnvPrefix("istio")
	viper.AutomaticEnv() // read in environment variables that match

	// Allow root flags to be specified in ENV or config file.
	viper.BindPFlags(RootCmd.LocalFlags())
	viper.BindPFlags(RootCmd.PersistentFlags())

	// Shut up glog log before parse
	flag.CommandLine.Parse([]string{})

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

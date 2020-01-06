/*
 * Copyright 2019-2020 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/option"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var operatorConfigFile string
var rootCmd = &cobra.Command{
	Use:   "dgraph-operator",
	Short: "Dgraph Operator creates/configures/manages Dgraph clusters atop Kubernetes.",

	Run: func(cmd *cobra.Command, args []string) {
		glog.Infof("starting dgraph operator")
		RunOperator()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootFlags := rootCmd.Flags()
	rootFlags.BoolVar(&option.OperatorConfig.SkipCRDCreation,
		"skip-crd", false, "Skip kubernetes custom resource definition creation.")
	rootFlags.StringVar(&operatorConfigFile,
		"config-file", "", "Configuration file. Takes precedence over default values, but is "+
			"overridden to values set with environment variables and flags.")
	rootFlags.StringVar(&option.OperatorConfig.KubeCfgPath, "kubecfg-path", "",
		"Path of kubeconfig file.")
	rootFlags.StringVar(&option.OperatorConfig.K8sAPIServerURL, "k8s-api-server-url", "",
		"URL of the kubernetes API server.")
	rootFlags.StringVar(&option.OperatorConfig.Server.Host, "server.host",
		defaults.OperatorHost, "Host to listen on.")
	rootFlags.IntVar(&option.OperatorConfig.Server.Port, "server.port",
		defaults.OperatorPort, "Port to listen on.")
	rootFlags.IntVar(&option.OperatorConfig.WorkersCount, "workers",
		defaults.WorkersCount, "Number of workers to run for the controller.")

	// Convinces glog that Parse() has been called to avoid noisy logs.
	// https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	if err := flag.CommandLine.Parse([]string{}); err != nil {
		glog.Fatalf("error while parsing command line flags: %s", err)
	}

	// Add all existing global flag (eg: from glog) to rootCmd's flags
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	if err := rootCmd.PersistentFlags().Set("stderrthreshold", "0"); err != nil {
		glog.Fatalf("error while setting default value for glog flag: %s", err)
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cmdRefCmd)

	if err := viper.BindPFlags(rootFlags); err != nil {
		glog.Fatalf("error while binding flag set to viper configuration: %s", err)
	}
	cobra.OnInitialize(initCmds)
}

func initCmds() {
	if operatorConfigFile == "" {
		return
	}

	viper.SetConfigFile(operatorConfigFile)
	viper.SetEnvPrefix("DGRAPH_OPERATOR")
	viper.AutomaticEnv()

	var err error
	if err = viper.ReadInConfig(); err == nil {
		glog.Infof("using config file: %s", viper.ConfigFileUsed())
	}

	setGlogFlags()
}

func setGlogFlags() {
	glogFlags := [...]string{
		"log_dir", "logtostderr", "alsologtostderr", "v",
		"stderrthreshold", "vmodule", "log_backtrace_at",
	}
	for _, gflag := range glogFlags {
		// Set value of flag to the value in config
		if stringValue, ok := viper.Get(gflag).(string); ok {
			if gflag == "log_backtrace_at" && stringValue == ":0" {
				continue
			}
			_ = flag.Lookup(gflag).Value.Set(stringValue)
		}
	}
}

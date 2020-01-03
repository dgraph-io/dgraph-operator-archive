package main

import (
	"fmt"
	"os"

	"github.com/dgraph-io/dgraph-operator/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the current build of dgraph operator.",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(version.VersionFormatStr,
			version.Info["apiVersion"],
			version.Info["operatorVersion"],
			version.Info["commitSHA"],
			version.Info["commitTimestamp"],
			version.Info["branch"],
			version.Info["goVersion"])
		os.Exit(0)
	},
}

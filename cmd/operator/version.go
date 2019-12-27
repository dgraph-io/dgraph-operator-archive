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
			version.Info["version"],
			version.Info["revision"],
			version.Info["branch"],
			version.Info["buildDate"],
			version.Info["goVersion"])
		os.Exit(0)
	},
}

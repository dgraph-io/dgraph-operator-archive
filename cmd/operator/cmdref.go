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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var cmdRefDir string

// cmdRefCmd is the operator command to generate command line reference docs using
// spf13/cobra/doc
var cmdRefCmd = &cobra.Command{
	Use:   "cmdref",
	Short: "Generate command line reference for dgraph operator command line interface.",

	Run: func(cmd *cobra.Command, args []string) {
		// Disable autogen comment in the command line reference.
		rootCmd.DisableAutoGenTag = true

		fmt.Printf("generating command line reference documentation in directory: %s\n", cmdRefDir)
		err := doc.GenMarkdownTree(rootCmd, cmdRefDir)
		if err != nil {
			fmt.Printf("error while generating cmdref: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cmdRefCmd.Flags().StringVarP(&cmdRefDir, "directory", "d",
		"docs/cmdref/", "Directory to use for creating cmd reference docs.")
}

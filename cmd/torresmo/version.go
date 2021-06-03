package main

import "github.com/spf13/cobra"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Torresmo version",
	Run: func(cmd *cobra.Command, args []string) {
		println(Version)
	},
}

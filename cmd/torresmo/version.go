package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Torresmo version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s (%s)", Version, Commit)
	},
}

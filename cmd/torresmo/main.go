package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mvrilo/torresmo"
	"github.com/spf13/cobra"
)

var Commit string
var Version string

func main() {
	runtime.LockOSThread()

	torresm, err := torresmo.New()
	if err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:     "torresmo",
		Short:   "Torresmo torrent client and server",
		Version: fmt.Sprintf("%s (%s)", Version, Commit),
	}

	rootCmd.AddCommand(serverCmd(torresm))
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/mvrilo/torresmo"
	"github.com/spf13/cobra"
)

func main() {
	runtime.SetMutexProfileFraction(1)
	runtime.LockOSThread()
	flag.Parse()

	rootCmd := &cobra.Command{
		Use:   "torresmo",
		Short: "Torresmo is an experimental torrent client",
		// Run: func(cmd *cobra.Command, args []string) {
		// },
	}

	ctx := context.Background()
	torresm, err := torresmo.New()
	if err != nil {
		panic(err)
	}

	rootCmd.AddCommand(serverCmd(ctx, torresm))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

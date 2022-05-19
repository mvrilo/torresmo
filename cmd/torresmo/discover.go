package main

import (
	"io"
	"log"
	"strings"

	"github.com/mvrilo/torresmo/mdns"
	"github.com/spf13/cobra"
)

func discoverCmd() *cobra.Command {
	log.SetOutput(io.Discard)
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover Torresmo servers in the network",
		Run: func(cmd *cobra.Command, args []string) {
			if svcs := mdns.SearchServices(); len(svcs) > 0 {
				println(strings.Join(svcs, "\n"))
			}
		},
	}
	return cmd
}

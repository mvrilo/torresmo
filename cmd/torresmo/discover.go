package main

import (
	"io"
	"log"
	"strings"

	"github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/mdns"
	"github.com/spf13/cobra"
)

func discoverCmd(torresm *torresmo.Torresmo) *cobra.Command {
	log.SetOutput(io.Discard)

	return &cobra.Command{
		Use:   "discover",
		Short: "Discover Torresmo servers in the network",
		Run: func(cmd *cobra.Command, args []string) {
			println(strings.Join(mdns.SearchServices(), "\n"))
		},
	}
}

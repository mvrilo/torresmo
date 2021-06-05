package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/mdns"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

func discoverCmd(torresm *torresmo.Torresmo) *cobra.Command {
	log.SetOutput(io.Discard)
	var openp bool

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover Torresmo servers in the network",
		Run: func(cmd *cobra.Command, args []string) {
			svcs := mdns.SearchServices()
			if len(svcs) == 0 {
				return
			}

			println(strings.Join(svcs, "\n"))
			if openp {
				addr := fmt.Sprintf("http://%s", strings.Split(svcs[0], " ")[0])
				open.Run(addr)
			}
		},
	}

	cmd.Flags().BoolVarP(&openp, "open", "o", false, "Open service address in the browser")
	return cmd
}

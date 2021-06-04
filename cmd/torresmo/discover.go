package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hashicorp/mdns"
	"github.com/mvrilo/torresmo"
	"github.com/spf13/cobra"
)

func discoverCmd(torresm *torresmo.Torresmo) *cobra.Command {
	log.SetOutput(io.Discard)

	return &cobra.Command{
		Use:   "discover",
		Short: "Discover Torresmo servers in the network",
		Run: func(cmd *cobra.Command, args []string) {
			entriesCh := make(chan *mdns.ServiceEntry, 4)
			go func() {
				for entry := range entriesCh {
					if !strings.ContainsAny(entry.Name, serviceName) {
						continue
					}

					fmt.Printf("%s:%d %s (%s)", entry.AddrV4, entry.Port, entry.Name, entry.AddrV6)
					if entry.Info != "" {
						fmt.Printf(" %s", entry.Info)
					}
					fmt.Println("")
				}
			}()

			mdns.Lookup(serviceName, entriesCh)
			close(entriesCh)
		},
	}
}

package mdns

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hashicorp/mdns"
)

// ServiceName describes a mdns service name
const ServiceName = "_torresmo._tcp"

func init() {
	log.SetOutput(io.Discard)
}

// SearchServices does a mdns lookup for torresmo services
func SearchServices() (entries []string) {
	entriesCh := make(chan *mdns.ServiceEntry, 4)

	go func() {
		for entry := range entriesCh {
			if !strings.ContainsAny(entry.Name, ServiceName) {
				continue
			}

			entries = append(entries, fmt.Sprintf(
				"%s:%d %s (%s) %s",
				entry.AddrV4,
				entry.Port,
				entry.Name,
				entry.AddrV6,
				entry.Info,
			))
		}
	}()

	mdns.Lookup(ServiceName, entriesCh)
	close(entriesCh)
	return
}

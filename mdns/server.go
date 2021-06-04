package mdns

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/mdns"
	"github.com/pkg/errors"
)

type Server struct {
	*mdns.Server
	Addr string
}

func Hostname() (host, fullhost string) {
	host, _ = os.Hostname()
	host = strings.Replace(host, ".local", "", -1)
	fullhost = fmt.Sprintf("%s.%s.local", host, ServiceName)
	return
}

func NewServer(addr string) (*Server, error) {
	srv := &Server{
		Addr: addr,
	}

	_, port, _ := net.SplitHostPort(addr)
	iport, _ := strconv.Atoi(port)

	host, _ := Hostname()
	service, err := mdns.NewMDNSService(host, ServiceName, "", "", iport, nil, []string{})
	if err != nil {
		return nil, errors.Wrap(err, "error initializing mDNS service")
	}

	// log.Info("Starting mDNS server with service name: ", fullhost)
	mdnsServer, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return nil, errors.Wrap(err, "error starting mDNS server")
	}

	srv.Server = mdnsServer
	return srv, nil
}

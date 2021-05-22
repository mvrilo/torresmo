package cast

import (
	"context"
	"errors"
	"net"
	"sync"

	castapp "github.com/vishen/go-chromecast/application"
	castdns "github.com/vishen/go-chromecast/dns"
)

var ErrMissingInterface = errors.New("missing network interface")

// Device holds a dns entry
type Device castdns.CastDNSEntry

// Cast is a Chromecast application mutex
type Cast struct {
	*castapp.Application
	mu               *sync.Mutex
	lastFoundDevices []Device

	// Network Interface
	Interface string

	// Device connected
	Device Device
}

func newapp() *castapp.Application {
	castappOptions := []castapp.ApplicationOption{
		castapp.WithDebug(false),
		castapp.WithCacheDisabled(true),
	}
	return castapp.NewApplication(castappOptions...)
}

// New initializes a new Cast
func New() *Cast {
	return &Cast{
		mu:          new(sync.Mutex),
		Application: newapp(),
		Device:      nil,
	}
}

func (c *Cast) discover(ctx context.Context) (<-chan castdns.CastEntry, error) {
	if c.Interface == "" {
		return nil, ErrMissingInterface
	}

	iface, err := net.InterfaceByName(c.Interface)
	if err != nil {
		return nil, err
	}

	return castdns.DiscoverCastDNSEntries(ctx, iface)
}

func (c *Cast) ListDevices(ctx context.Context) ([]Device, error) {
	devicesChan, err := c.discover(ctx)
	if err != nil {
		return nil, err
	}

	var devices []Device
	for device := range devicesChan {
		devices = append(devices, device)
	}

	c.lastFoundDevices = devices
	return devices, nil
}

func (c *Cast) Connect(ctx context.Context, uuid string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var device Device
	devices := c.lastFoundDevices
	if devices != nil && len(devices) > 1 {
		for _, d := range devices {
			if d.GetUUID() == uuid {
				device = d
				break
			}
		}
	}

	if device == nil {
		devicesChan, err := c.discover(ctx)
		if err != nil {
			return err
		}

		for dev := range devicesChan {
			if dev.GetUUID() == uuid {
				device = dev
				break
			}
			devices = append(devices, device)
		}
	}

	c.Device = device
	addr := device.GetAddr()
	port := device.GetPort()
	return c.Application.Start(addr, port)
}

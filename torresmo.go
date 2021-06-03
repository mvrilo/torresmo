package torresmo

import (
	"context"
	"embed"
	"io/fs"
	gohttp "net/http"
	"time"

	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/stream"
	"github.com/mvrilo/torresmo/torrent"
)

//go:embed static/dist
var staticFiles embed.FS

type Torresmo struct {
	HTTPServer    *gohttp.Server
	HTTPHandler   gohttp.Handler
	Publisher     stream.Publisher
	TorrentClient torrent.Client
	Logger        log.Logger
	StaticFiles   fs.FS
}

func (t *Torresmo) Shutdown(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Info("Shutting down torrent client")
	t.TorrentClient.Stop()

	if t.HTTPServer != nil {
		log.Info("Shutting down http server")

		if err := t.HTTPServer.Shutdown(ctx); err != nil && err != gohttp.ErrServerClosed {
			return err
		}
	}

	return nil
}

// New initializes a Torresmo with some defaults
func New() (*Torresmo, error) {
	logger := log.NewLogger()
	publisher := stream.NewWebsocket(logger)

	staticFiles, err := fs.Sub(staticFiles, "static/dist")
	if err != nil {
		return nil, err
	}

	torrentClient, err := torrent.NewClient(logger)
	if err != nil {
		return nil, err
	}

	return &Torresmo{
		Logger:        logger,
		Publisher:     publisher,
		TorrentClient: torrentClient,
		StaticFiles:   staticFiles,
	}, nil
}

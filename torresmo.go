package torresmo

import (
	"context"
	"io/fs"
	gohttp "net/http"
	"time"

	"github.com/mvrilo/torresmo/event"
	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/static"
	"github.com/mvrilo/torresmo/torrent"
)

type Torresmo struct {
	HTTPServer    *gohttp.Server
	HTTPHandler   gohttp.Handler
	EventHandler  event.Handler
	TorrentClient torrent.Client
	Logger        log.Logger
	StaticFiles   fs.FS
}

func (t *Torresmo) Shutdown(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if t.HTTPServer != nil {
		log.Info("Shutting down http server")

		if err := t.HTTPServer.Shutdown(ctx); err != nil && err != gohttp.ErrServerClosed {
			return err
		}
	}

	log.Info("Shutting down torrent client")
	t.TorrentClient.Stop()

	return nil
}

// New initializes a Torresmo with some defaults
func New() (*Torresmo, error) {
	logger := log.NewLogger()
	wshandler := event.NewWebsocket(logger)

	staticFiles, err := fs.Sub(static.Files, "dist")
	if err != nil {
		return nil, err
	}

	torrentClient, err := torrent.NewClient(logger)
	if err != nil {
		return nil, err
	}

	return &Torresmo{
		Logger:        logger,
		EventHandler:  wshandler,
		TorrentClient: torrentClient,
		StaticFiles:   staticFiles,
	}, nil
}

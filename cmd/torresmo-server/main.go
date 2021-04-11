package main

import (
	"context"
	"flag"
	"fmt"
	gohttp "net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/gui"
	"github.com/mvrilo/torresmo/http"
	"github.com/mvrilo/torresmo/log"
	"github.com/spf13/cobra"
)

func main() {
	runtime.LockOSThread()
	flag.Parse()

	rootCmd := &cobra.Command{
		Use:   "torresmo-server",
		Short: "Torresmo is an experimental torrent client",
		Run: func(cmd *cobra.Command, args []string) {
			os.MkdirAll("/tmp/torresmo", 0750)

			ctx := context.Background()
			torresm, err := torresmo.New()
			if err != nil {
				panic(err)
			}

			errCh := make(chan error)
			torresm.Logger.Info(fmt.Sprintf("Starting HTTP server at localhost:8000"))

			torresm.HTTPHandler = http.NewHandler(
				torresm.TorrentClient,
				torresm.Logger,
				torresm.StaticFiles,
				torresm.Publisher,
				false,
			)

			torresm.HTTPServer = &gohttp.Server{
				Addr:    "127.0.0.1:8000",
				Handler: torresm.HTTPHandler,
			}
			go func() {
				if err := torresm.HTTPServer.ListenAndServe(); err != gohttp.ErrServerClosed {
					log.Error("Error starting HTTP server:", err)
				}
			}()

			var guiApp gui.GUI
			go func(ctx context.Context, torresm *torresmo.Torresmo, errCh chan error) {
				sig := make(chan os.Signal, 1)
				signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
				<-sig

				if err := torresm.Shutdown(ctx, 10*time.Second); err != nil {
					errCh <- err
				}
				close(errCh)

				if guiApp != nil {
					guiApp.Stop()
				}
			}(ctx, torresm, errCh)

			cli := torresm.
				TorrentClient.
				WithPublisher(torresm.Publisher).
				WithOutput("/tmp/torresmo/downloads").
				// WithDownloadLimit(downloadLimit).
				// WithUploadLimit(uploadLimit).
				WithSeed(true).
				WithWatchDir("/tmp/torresmo/downloads").
				WithBiggestFirst(true)

			err = cli.Start()
			if err != nil {
				panic(err)
			}

			err = cli.ReadTorrentFiles()
			if err != nil {
				panic(err)
			}

			guiApp = gui.NewGUI(torresm)
			guiApp.Start()

			if err := <-errCh; err != nil {
				log.Error("Error shutting down server:", err)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

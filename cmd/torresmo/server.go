package main

import (
	"context"
	"fmt"
	gohttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/gui"
	"github.com/mvrilo/torresmo/http"
	"github.com/mvrilo/torresmo/log"
	"github.com/spf13/cobra"
)

func serverCmd(torresm *torresmo.Torresmo) *cobra.Command {
	var out string
	var addr string
	var watchDir string
	var debug bool
	var guiFlag bool
	var seedFlag bool
	var biggestFirst bool
	var uploadLimit int
	var downloadLimit int
	var castIface string

	srvCmd := &cobra.Command{
		Use:   "server",
		Short: "Torresmo torrent client and server",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			errCh := make(chan error)

			if addr != "" {
				torresm.Logger.Info(fmt.Sprintf("Starting HTTP server at %s", addr))

				torresm.HTTPHandler = http.NewHandler(
					torresm.TorrentClient,
					torresm.Logger,
					torresm.StaticFiles,
					torresm.Publisher,
					torresm.Cast,
					debug,
				)

				if castIface != "" {
					torresm.Cast.Interface = castIface
				}

				torresm.HTTPServer = &gohttp.Server{
					Addr:    addr,
					Handler: torresm.HTTPHandler,
				}
				go func() {
					if err := torresm.HTTPServer.ListenAndServe(); err != gohttp.ErrServerClosed {
						log.Error("Error starting HTTP server:", err)
					}
				}()
			}

			var guiApp gui.GUI
			go func() {
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
			}()

			cli := torresm.
				TorrentClient.
				WithPublisher(torresm.Publisher).
				WithOutput(out).
				WithDownloadLimit(downloadLimit).
				WithUploadLimit(uploadLimit).
				WithSeed(seedFlag).
				WithWatchDir(watchDir).
				WithBiggestFirst(biggestFirst)

			err := cli.Start()
			if err != nil {
				panic(err)
			}

			if watchDir != "" {
				err := cli.ReadTorrentFiles()
				if err != nil {
					panic(err)
				}
			}

			if guiFlag {
				guiApp = gui.NewGUI(torresm)
				guiApp.Start()
			}

			if err := <-errCh; err != nil {
				log.Error("Error shutting down server:", err)
			}
		},
	}

	srvCmd.Flags().StringVarP(&out, "out", "o", "downloads", "Output directory")
	srvCmd.Flags().BoolVarP(&guiFlag, "gui", "g", true, "Runs graphical interface")
	srvCmd.Flags().BoolVarP(&seedFlag, "seed", "s", true, "Enable seeding")
	srvCmd.Flags().BoolVarP(&debug, "debug", "d", true, "Enable seeding")
	srvCmd.Flags().BoolVarP(&biggestFirst, "biggest", "b", true, "Prioritize the biggest file in the torrent")
	srvCmd.Flags().StringVarP(&castIface, "castiface", "c", "", "Interface enabled Chromecast interaction, eg: eth0, en1")
	srvCmd.Flags().StringVarP(&watchDir, "watch", "w", "downloads", "Watch torrents in this directory")
	srvCmd.Flags().IntVarP(&uploadLimit, "upload-limit", "U", 0, "Upload limit")
	srvCmd.Flags().IntVarP(&downloadLimit, "download-limit", "D", 0, "Download limit")
	srvCmd.Flags().StringVarP(&addr, "addr", "a", ":8000", "HTTP Server address")
	return srvCmd
}

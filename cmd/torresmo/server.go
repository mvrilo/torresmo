package main

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	gohttp "net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/mdns"
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
	var serve bool
	var biggestFirst bool
	var uploadLimit int
	var downloadLimit int
	var enableDiscovery bool

	srvCmd := &cobra.Command{
		Use:   "server",
		Short: "Torresmo torrent client and server",
		Run: func(cmd *cobra.Command, args []string) {
			if addr == "" {
				log.Error("Missing server address")
				os.Exit(1)
			}

			ctx := context.Background()
			torresm.Logger.Info(fmt.Sprintf("Starting HTTP server at %s", addr))
			var outFiles fs.FS
			if serve {
				outFiles = os.DirFS(out)
			}

			torresm.HTTPHandler = http.NewHandler(
				torresm.TorrentClient,
				torresm.Logger,
				torresm.StaticFiles,
				outFiles,
				torresm.Publisher,
				debug,
			)

			torresm.HTTPServer = &gohttp.Server{
				Addr:    addr,
				Handler: torresm.HTTPHandler,
			}

			go func() {
				if err := torresm.HTTPServer.ListenAndServe(); err != gohttp.ErrServerClosed {
					log.Error("Error starting HTTP server:", err)
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
				log.Error("Error starting Torrent client:", err)
				os.Exit(1)
			}

			if watchDir != "" {
				err := cli.ReadTorrentFiles()
				if err != nil {
					log.Error("Error reading torrent files:", err)
					os.Exit(1)
				}
			}

			var mdnsServer *mdns.Server
			if enableDiscovery {
				_, port, _ := net.SplitHostPort(addr)
				iport, _ := strconv.Atoi(port)

				host, _ := os.Hostname()
				host = strings.Replace(host, ".local", "", -1)
				service, _ := mdns.NewMDNSService(host, mdnsServiceName, "", "", iport, nil, []string{})

				log.Info("Starting mDNS server as: ", host+"."+mdnsServiceName+".local")
				mdnsServer, err = mdns.NewServer(&mdns.Config{Zone: service})
				if err != nil {
					log.Error("Error starting mdns server:", err)
				} else {
					defer mdnsServer.Shutdown()
				}
			}

			if guiFlag && gui.App != nil {
				gui.App.Register(torresm)
				gui.App.Start()
				defer gui.App.Stop()
			}

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig

			if err := torresm.Shutdown(ctx, 10*time.Second); err != nil {
				log.Error("Error shutting down server:", err)
			}
		},
	}

	srvCmd.Flags().StringVarP(&out, "out", "o", "downloads", "Output directory")
	srvCmd.Flags().BoolVarP(&guiFlag, "gui", "g", true, "Runs graphical interface")
	srvCmd.Flags().BoolVarP(&seedFlag, "seed", "s", true, "Enable seeding")
	srvCmd.Flags().BoolVarP(&debug, "debug", "d", true, "Enable seeding")
	srvCmd.Flags().BoolVarP(&serve, "serve", "e", true, "Serve downloaded files")
	srvCmd.Flags().BoolVarP(&biggestFirst, "biggest", "b", true, "Prioritize the biggest file in the torrent")
	srvCmd.Flags().BoolVarP(&enableDiscovery, "discovery", "c", true, "Enable mDNS discovery")
	srvCmd.Flags().StringVarP(&watchDir, "watch", "w", "downloads", "Watch torrents in this directory")
	srvCmd.Flags().IntVarP(&uploadLimit, "upload-limit", "U", 0, "Upload limit")
	srvCmd.Flags().IntVarP(&downloadLimit, "download-limit", "D", 0, "Download limit")
	srvCmd.Flags().StringVarP(&addr, "addr", "a", ":8000", "HTTP Server address")
	return srvCmd
}

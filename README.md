# Torresmo

<!--[![GoDoc](https://godoc.org/github.com/mvrilo/protog?status.svg)](https://godoc.org/github.com/mvrilo/protog) 
[![Go Report Card](https://goreportcard.com/badge/github.com/mvrilo/protog)](https://goreportcard.com/report/github.com/mvrilo/protog) -->

Torresmo is an experimental and tasty torrent client.

## Features

- single binary
- rest api
- websocket events
- embedded web interface (with typescript and preact)
- graphical interface (initial support, mac only)

## Usage

```
$ ./torresmo server -h
Torresmo's Torrent and HTTP server

Usage:
  torresmo server [flags]

Flags:
  -a, --addr string           HTTP Server address (default ":8000")
  -b, --biggestfirst          Prioritize the biggest file in the torrent (default true)
  -d, --debug                 Enable seeding (default true)
  -D, --download-limit int    Download limit
  -g, --gui                   Runs graphical interface (default true)
  -h, --help                  help for server
  -o, --out string            Output directory (default "downloads")
  -s, --seed                  Enable seeding (default true)
  -t, --torrentfiles string   Read torrent files from directory (default "downloads")
  -U, --upload-limit int      Upload limit
  -w, --watch string          Watch torrents in this directory (default "downloads")
```

## Example

![Example](demo.png)

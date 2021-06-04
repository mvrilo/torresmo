# Torresmo

[![GoDoc](https://godoc.org/github.com/mvrilo/torresmo?status.svg)](https://godoc.org/github.com/mvrilo/torresmo)
[![Go Report Card](https://goreportcard.com/badge/github.com/mvrilo/torresmo)](https://goreportcard.com/report/github.com/mvrilo/torresmo)

Torresmo is an experimental Torrent client and server built with Go.

## Features

- single binary
- http server and rest api
- websocket events and rooms
- embedded web interface (using esbuild, typescript and preact)
- graphical interface (mac only for now, via macdriver)
- mdns discovery mechanism

## Build

Go binary:

```
make torresmo
```

## Usage

Torresmo

```
Torresmo torrent client and server

Usage:
  torresmo [command]

Available Commands:
  discover    Discover Torresmo servers in the network
  help        Help about any command
  server      Torresmo torrent client and server
  version     Torresmo version

Flags:
  -h, --help      help for torresmo
  -v, --version   version for torresmo
```

Server usage:

```
Torresmo torrent client and server

Usage:
  torresmo server [flags]

Flags:
  -a, --addr string          HTTP Server address (default ":8000")
  -b, --biggest              Prioritize the biggest file in the torrent (default true)
  -d, --debug                Enable seeding (default true)
  -c, --discovery            Enable mdns discovery (default true)
  -D, --download-limit int   Download limit
  -g, --gui                  Runs graphical interface (default true)
  -h, --help                 help for server
  -o, --out string           Output directory (default "downloads")
  -s, --seed                 Enable seeding (default true)
  -e, --serve                Serve downloaded files (default true)
  -U, --upload-limit int     Upload limit
  -w, --watch string         Watch torrents in this directory (default "downloads")
```

## Example

![Example](demo.png)

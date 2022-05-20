.PHONY: all run dev debug web prepare clean test release build build-dev check

COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS = -X main.Commit=$(COMMIT) -X main.Version=$(VERSION)

all: torresmo

torresmo: prepare static/dist/bundle.js
	time go build -ldflags="-s -w $(LDFLAGS)" -o torresmo cmd/torresmo/*.go

torresmo-dev: web
	time go build -ldflags="$(LDFLAGS)" -race -o torresmo-dev cmd/torresmo/*.go

build: torresmo

build-dev: torresmo-dev

run: torresmo
	./torresmo server --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=900 --download-limit=90000

dev: torresmo-dev
	./torresmo-dev server --gui --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=900 --download-limit=90000

debug: torresmo-dev
	./torresmo-dev server --debug --gui --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=900 --download-limit=90000

web: static/dist/bundle.js

static/dist/bundle.js:
	@(cd static; yarn build)

check:
	go fmt ./...
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck -- $$(go list ./...)

prepare: check
	go mod tidy

test:
	go test ./...

clean:
	rm -rf torresmo torresmo-dev static/dist/bundle.js dist/* 2>/dev/null

release:
	/bin/sh release.sh

.PHONY: all run dev debug mac web prepare clean test release macapp

COMMIT = $(shell git rev-parse --short HEAD)
VERSION = $(shell cat version)
LDFLAGS = -X main.Commit=$(COMMIT) -X main.Version=$(VERSION)
GCFLAGS = -l -m=2 -d=checkptr
GODEBUG = cgocheck=2

all: torresmo

torresmo: prepare static/dist/bundle.js
	time go build -ldflags="-s -w $(LDFLAGS)" -o torresmo cmd/torresmo/*.go

torresmo-dev: web
	time go build -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" -race -o torresmo-dev cmd/torresmo/*.go

run: torresmo
	./torresmo server --gui --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=100 --download-limit=9000

dev: torresmo-dev
	GODEBUG=$(GODEBUG) ./torresmo-dev server --gui --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=100 --download-limit=90

debug: torresmo-dev
	GODEBUG=$(GODEBUG) ./torresmo-dev server --debug --gui --discovery --serve --out=downloads --torrent-files=downloads/.torrents --addr=:8000 --upload-limit=100 --download-limit=500000

tools/macapp/macapp:
	time go build -o ./tools/macapp/macapp ./tools/macapp/main.go

macapp: tools/macapp/macapp torresmo
	cp torresmo assets/;
	./tools/macapp/macapp \
		-assets=./assets \
		-bin=torresmo-mac.sh \
		-dmg=Torresmo \
		-name=Torresmo \
		-o=./dist \
		-identifier co.murilo.torresmo \
		-icon=./assets/icon.png

web: static/dist/bundle.js

static/dist/bundle.js:
	@(cd static; yarn build)

prepare:
	go mod tidy
	go fmt ./...
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck -- $$(go list ./...)

test:
	go test ./...

clean:
	rm -rf torresmo torresmo-dev static/dist/bundle.js dist/* 2>/dev/null

release:
	git tag v$(VERSION)
	git push origin v$(VERSION)
	go run github.com/goreleaser/goreleaser release --rm-dist

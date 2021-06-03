.PHONY: all run dev debug mac web prepare clean test release

COMMIT = $(shell git rev-parse --short HEAD)
VERSION = $(shell cat version)
LDFLAGS = -X main.Commit=$(COMMIT) -X main.Version=$(VERSION)

all: torresmo

torresmo:
	time go build -ldflags="-s -w $(LDFLAGS)" -o torresmo cmd/torresmo/*.go

torresmo-dev: prepare
	time go build -ldflags="$(LDFLAGS)" -race -o torresmo-dev cmd/torresmo/*.go

run: torresmo
	./torresmo server --gui --serve --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=9000

dev: torresmo-dev
	./torresmo-dev server --gui --serve --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=90

debug: torresmo
	./torresmo server --debug --gui --serve --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=500000

macapp:
	go build -o macapp ./tools/macapp/main.go

mac: macapp torresmo
	chmod +x macassets/torresmo.sh;
	cp torresmo macassets/;
	./macapp \
		-assets=./macassets \
		-bin=torresmo.sh \
		-dmg=Torresmo \
		-name=Torresmo \
		-o=./dist \
		-identifier co.murilo.torresmo \
		-icon=./macassets/icon.png

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

build: clean prepare
	go run github.com/goreleaser/goreleaser build --rm-dist

release: build
	git tag v$(VERSION)
	git push origin v$(VERSION)
	go run github.com/goreleaser/goreleaser release --rm-dist

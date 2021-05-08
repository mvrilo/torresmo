all: torresmo

debug: torresmo
	./torresmo server --debug --gui --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=500000

run: torresmo
	./torresmo server --gui --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=9000

dev: torresmo-dev
	./torresmo-dev server --gui --watch=downloads --out=downloads --addr=:8000 --upload-limit=100 --download-limit=90

torresmo:
	time go build -ldflags="-s -w" -o torresmo cmd/torresmo/*.go

torresmo-dev: prepare
	time go build -race -o torresmo-dev cmd/torresmo/*.go

torresmo-server: prepare
	time go build -o torresmo-server cmd/torresmo-server/main.go

mac: macapp torresmo-server
	cp torresmo-server assets/;
	./macapp \
		-assets=./assets \
		-bin=torresmo-server \
		-dmg=Torresmo \
		-name=Torresmo \
		-o=./dist \
		-identifier co.murilo.torresmo \
		-icon=./assets/icon.png

macapp:
	go build -o macapp ./tools/macapp/main.go

web: static/dist/bundle.js

static/dist/bundle.js:
	@(cd static; yarn build)

prepare:
	go fmt ./...
	go vet ./...

test:
	go test ./...

clean:
	rm -rf torresmo torresmo-dev static/dist/bundle.js 2>/dev/null

module github.com/mvrilo/torresmo

go 1.16

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/DeanThompson/ginpprof v0.0.0-20201112072838-007b1e56b2e1
	github.com/anacrolix/log v0.8.0
	github.com/anacrolix/torrent v1.25.1-0.20210224024805-693c30dd889e
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.0.4
	github.com/json-iterator/go v1.1.10
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/progrium/macdriver v0.1.0
	github.com/spf13/cobra v1.1.3
	github.com/ugorji/go v1.2.4 // indirect
	github.com/vishen/go-chromecast v0.2.9
	go.uber.org/zap v1.16.0
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/tools v0.1.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.6 // indirect
)

replace github.com/progrium/macdriver => github.com/mvrilo/macdriver v0.1.1-0.20210407101456-ec21e5ee45f9

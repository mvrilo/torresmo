//go:build tools
// +build tools

package tools

import (
	_ "github.com/evanw/esbuild/cmd/esbuild"
	_ "github.com/goreleaser/goreleaser"
	_ "honnef.co/go/tools/cmd/staticcheck"
)

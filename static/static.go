package static

import "embed"

//go:embed dist
// Files is the embedded static files
var Files embed.FS

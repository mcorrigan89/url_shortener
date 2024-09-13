package public

import "embed"

//go:embed "static" "favicon.ico" "robots.txt"
var Files embed.FS

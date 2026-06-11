package embedfs

import "embed"

//go:embed dist
var FrontendFS embed.FS

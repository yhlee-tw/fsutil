package fsutiltest

import (
	"embed"
)

// TestEmbeddedFS is our testing go:embed FS.
//
//go:embed test_data.txt
var TestEmbeddedFS embed.FS

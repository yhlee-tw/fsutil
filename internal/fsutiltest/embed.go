package fsutiltest

import (
	"embed"
)

//go:embed test_data.txt
var TestEmbeddedFS embed.FS

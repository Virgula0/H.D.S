package static

import "embed"

//go:embed styles
var StylesFS embed.FS

//go:embed scripts
var ScriptsFS embed.FS

//go:embed images
var ImageFS embed.FS

package assets

import "embed"

//go:embed css
var CSS embed.FS

//go:embed images
var Images embed.FS

//go:embed js
var JS embed.FS

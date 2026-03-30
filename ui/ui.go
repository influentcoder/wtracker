// Package ui embeds the frontend static files.
package ui

import "embed"

//go:embed static
var StaticFS embed.FS

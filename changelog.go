// Package ytac embeds repository-level files used by the ytac binary.
package ytac

import _ "embed"

// Changelog is the contents of CHANGELOG.md.
//
//go:embed CHANGELOG.md
var Changelog string

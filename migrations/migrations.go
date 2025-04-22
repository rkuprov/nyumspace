package migrations

import "embed"

// Embedding the exported var here to be avaiable as an import in the daemon package.
//
//go:embed *.sql
var EmbeddedMigrations embed.FS

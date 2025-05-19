package migrations

import "embed"

//go:embed *.sql

// Embedded contains the SQL migration files.
var Embedded embed.FS

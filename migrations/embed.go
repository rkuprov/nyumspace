package migrations

import "embed"

//go:embed *.sql

// Embedded contains the SQL migration files.
// these are needed for the integration tests
var Embedded embed.FS

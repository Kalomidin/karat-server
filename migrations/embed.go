// Package migrations holds the SQL that creates the Karat database.
package migrations

import "embed"

// Files is every migration, applied in name order.
//
//go:embed *.sql
var Files embed.FS

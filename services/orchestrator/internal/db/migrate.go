package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies incremental SQL (002+) that is safe to re-run.
// 001 is docker-entrypoint / first boot only.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if strings.HasPrefix(e.Name(), "001_") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		body, err := fs.ReadFile(migrationFS, "migrations/"+name)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		log.Printf("applied %s", name)
	}
	return nil
}

package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/migrations"
)

// RunMigrations reads embedded *.up.sql files and applies them in order.
// A schema_migrations table tracks which versions have been applied.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := make(map[int]bool)
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan migration version: %w", err)
		}
		applied[v] = true
	}

	type migrationFile struct {
		version int
		name    string
		content string
	}

	var files []migrationFile
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(name, "%d_", &version); err != nil {
			continue
		}

		if applied[version] {
			continue
		}

		data, err := migrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		files = append(files, migrationFile{
			version: version,
			name:    name,
			content: string(data),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	for _, f := range files {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", f.name, err)
		}

		if _, err := tx.Exec(ctx, f.content); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", f.name, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, f.version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", f.name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", f.name, err)
		}

		fmt.Printf("  migration %d applied: %s\n", f.version, f.name)
	}

	if len(files) == 0 {
		fmt.Println("  migrations already up to date")
	}

	return nil
}

package database

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// RunMigrations applies SQL schema if found (schema.sql in repo root or current cwd).
// This is a simple runner to ensure DB-backed state without requiring external tools.
func RunMigrations(db *gorm.DB) error {
	paths := []string{
		"schema.sql",
		"./schema.sql",
		"../schema.sql",
		"../../schema.sql",
	}

	var sqlBytes []byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			sqlBytes = b
			break
		}
	}

	// If schema not found, skip without error (ops scripts may have applied it)
	if len(sqlBytes) == 0 {
		return nil
	}

	statements := splitSQLStatements(string(sqlBytes))
	ctx := context.Background()

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
		}
		return nil
	})
}

func splitSQLStatements(sql string) []string {
	// naive split by semicolon; good enough for our single-file schema
	parts := strings.Split(sql, ";")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		trim := strings.TrimSpace(p)
		if trim != "" {
			clean = append(clean, trim)
		}
	}
	return clean
}

// DiscoverRepoRoot tries to find repository root (unused now, kept for future enhancement)
func DiscoverRepoRoot() (string, error) {
	start, _ := os.Getwd()
	return findUp(start, ".git")
}

func findUp(dir, needle string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, needle)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fs.ErrNotExist
		}
		dir = parent
	}
}

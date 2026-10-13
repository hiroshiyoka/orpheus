package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path, migrationPath string) (*sql.DB, error) {
	if directory := filepath.Dir(path); directory != "." {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000`); err != nil {
		db.Close()
		return nil, err
	}
	dir := migrationPath
	if info, err := os.Stat(migrationPath); err == nil && !info.IsDir() {
		dir = filepath.Dir(migrationPath)
	}
	migrations, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		db.Close()
		return nil, err
	}
	sort.Strings(migrations)
	if len(migrations) == 0 {
		migrations = []string{migrationPath}
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		db.Close()
		return nil, err
	}
	for _, m := range migrations {
		name := filepath.Base(m)
		var applied int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&applied); err != nil {
			db.Close()
			return nil, err
		}
		if applied > 0 {
			continue
		}
		data, err := os.ReadFile(m)
		if err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec(string(data)); err != nil {
			// ponytail: tolerate re-running ADD COLUMN on DBs created before tracking existed
			if !strings.Contains(err.Error(), "duplicate column name") {
				db.Close()
				return nil, err
			}
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			db.Close()
			return nil, err
		}
	}
	return db, nil
}

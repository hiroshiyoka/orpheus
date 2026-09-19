package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"

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
	for _, m := range migrations {
		data, err := os.ReadFile(m)
		if err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec(string(data)); err != nil {
			db.Close()
			return nil, err
		}
	}
	return db, nil
}

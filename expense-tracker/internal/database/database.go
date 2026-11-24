// Package database открывает SQLite и применяет SQL-миграции из папки migrations.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // драйвер SQLite на чистом Go (без CGO)
)

// Open открывает базу. DSN = путь к файлу + PRAGMA для каждого соединения:
// foreign_keys(1) — включает проверку внешних ключей,
// busy_timeout(5000) — ждать до 5 c, если база занята другой записью.
func Open(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Migrate применяет по порядку все *.sql файлы, которые ещё не были применены.
// Применённые версии запоминаются в таблице schema_migrations.
func Migrate(db *sql.DB, dir string) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("в папке %q нет файлов миграций (запускайте сервер из корня проекта)", dir)
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)

		var applied int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		for _, stmt := range splitStatements(string(content)) {
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("миграция %s: %w", version, err)
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Printf("применена миграция %s", version)
	}
	return nil
}

// splitStatements убирает строки-комментарии и делит скрипт на отдельные запросы.
func splitStatements(script string) []string {
	var lines []string
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		lines = append(lines, line)
	}
	var stmts []string
	for _, part := range strings.Split(strings.Join(lines, "\n"), ";") {
		if s := strings.TrimSpace(part); s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

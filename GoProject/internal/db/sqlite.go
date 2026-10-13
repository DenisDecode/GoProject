package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite" // драйвер SQLite на чистом Go
)

func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Читаем и выполняем файл миграции schema.sql
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}

	return db, nil
}

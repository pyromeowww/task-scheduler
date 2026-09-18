package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// schema — SQL-запрос для создания таблицы задач.
const schema = `CREATE TABLE IF NOT EXISTS scheduler (
id INTEGER PRIMARY KEY AUTOINCREMENT,
date CHAR(8) NOT NULL DEFAULT "",
title VARCHAR(128) NOT NULL DEFAULT "",
comment TEXT NOT NULL DEFAULT "",
repeat VARCHAR(128) NOT NULL DEFAULT ""
);`

// indexData — индекс по дате для ускорения выборки задач.
const indexData = `CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date)`

var (
	// DB — глобальное подключение к базе данных.
	DB              *sql.DB
	ErrTaskNotFound = errors.New("task not found")
)

// Close закрывает соединение с базой данных.
func Close() error {
	if DB == nil {
		return nil
	}
	err := DB.Close()
	DB = nil
	return err
}

// Init открывает базу данных и, если файл ещё не существовал,
// создаёт таблицу задач и индекс.
func Init(dbfile string) (err error) {

	dir := filepath.Dir(dbfile)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create db directory %q: %w", dir, err)
		}
	}

	var install bool
	// Проверяем, существует ли файл базы данных.
	info, err := os.Stat(dbfile)
	switch {
	case err == nil:
		// Файл есть: убеждаемся, что это не директория.
		if info.IsDir() {
			return fmt.Errorf("db path %q is a directory, not a file", dbfile)
		}
	case errors.Is(err, os.ErrNotExist):
		// Файла нет — устанавливаем флаг создания схемы.
		install = true
	default:
		// Любая другая ошибка.
		return err
	}

	defer func() {
		if err != nil {
			_ = Close()
		}
	}()

	// Открываем соединение с SQLite-базой.
	DB, err = sql.Open("sqlite", dbfile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	// Проверяем, что подключение действительно работает.
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Если база создана с нуля, добавляем таблицу и индекс.
	if install {
		if _, err = DB.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
		if _, err = DB.Exec(indexData); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

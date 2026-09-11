package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// schema — SQL-запрос для создания таблицы задач.
const schema = `CREATE TABLE scheduler (
id INTEGER PRIMARY KEY AUTOINCREMENT,
date CHAR(8) NOT NULL DEFAULT "",
title VARCHAR(128) NOT NULL DEFAULT "",
comment TEXT NOT NULL DEFAULT "",
repeat VARCHAR(128) NOT NULL DEFAULT ""
);`

// indexData — индекс по дате для ускорения выборки задач.
const indexData = `CREATE INDEX scheduler_date ON scheduler(date);`

var (
	// Db — глобальное подключение к базе данных.
	Db *sql.DB
	ErrTaskNotFound = errors.New("Задача не найдена")
)

// Init открывает базу данных и, если файл ещё не существовал,
// создаёт таблицу задач и индекс.
func Init(dbfile string) error {
	var install bool
	// Проверяем, существует ли файл базы данных.
	info, err := os.Stat(dbfile)
	switch {
	case err == nil:
		// Файл есть: убеждаемся, что это не директория.
		if info.IsDir() {
			return fmt.Errorf("db path %q is a directory, not a file", dbfile)
		}
	case os.IsNotExist(err):
		// Файла нет — устанавливаем флаг создания схемы.
		install = true
	default:
		// Любая другая ошибка.
		return err
	}

	// Открываем соединение с SQLite-базой.
	Db, err = sql.Open("sqlite", dbfile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	// Проверяем, что подключение действительно работает.
	if err := Db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Если база создана с нуля, добавляем таблицу и индекс.
	if install {
		if _, err := Db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
		if _, err := Db.Exec(indexData); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

// Close закрывает соединение с базой данных.
func Close() error {
	if Db != nil {
		return Db.Close()
	}
	return nil
}

package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const schema = `CREATE TABLE scheduler (
id INTEGER PRIMARY KEY AUTOINCREMENT,
date CHAR(8) NOT NULL DEFAULT "",
title VARCHAR(128) NOT NULL DEFAULT "",
comment TEXT NOT NULL DEFAULT "",
repeat VARCHAR(128) NOT NULL DEFAULT ""
);`
const indexData = `CREATE INDEX scheduler_date ON scheduler(date);`

// глобальная переменная, хранящая подключение к базе данных
var Db *sql.DB

// Init — открывает базу данных и при необходимости создавёт таблицу с индексом
func Init(dbfile string) error {
	var install bool
	// Провереряем есть ли файл
	info, err := os.Stat(dbfile)
	switch {
	case err == nil:
		// Файл есть, проверка, что это не директория
		if info.IsDir() {
			return fmt.Errorf("db path %q is a directory, not a file", dbfile)
		}
	case os.IsNotExist(err):
		// Файла нет, ставим флаг, что нужно создать
		install = true
	default:
		// Другая ошибка
		return err
	}

	Db, err = sql.Open("sqlite", dbfile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	// Проверяем подключение к БД
	if err := Db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

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

func Close() error {
	if Db != nil {
		return Db.Close()
	}
	return nil
}

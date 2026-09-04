
package main

import (
	"log"
	"os"

	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/server"
)


func main() {
	// Путь к БД берём из переменной окружения TODO_DBFILE.
	// Если она не задана, используем стандартное имя файла в рабочей папке.
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// Открываем базу данных и подготавливаем её к работе.
	// При ошибке программа завершается, так как без БД сервер не сможет работать.
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}

	// Отложенное закрытие базы данных.
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()
	
	server.RunServer()
}

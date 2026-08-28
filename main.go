package main

import (
	"log"
	"os"

	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/server"
)

func main() {
	// Пытаемся определить путь к БД через переменную окружения
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	// Инициализируем БД
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}
	// Закрываем базу при завершении всей программы
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()
	// Запуск сервера
	server.RunServer()
}

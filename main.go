package main

import (
	"log"

	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/server"
)

func main() {
	// Инициализируем БД
	if err := db.Init("scheduler.db"); err != nil {
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

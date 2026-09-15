package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/server"
	"github.com/pyromeowww/task-scheduler/pkg/settings"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, использую переменные окружения")
	}

	if os.Getenv(settings.EnvTodoPassword) != "" && os.Getenv(settings.EnvSaltJWT) == "" {
		log.Fatal("SALT_JWT не задан. Добавь его в .env или переменные окружения")
	}

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

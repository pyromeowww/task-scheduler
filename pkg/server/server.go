package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/api"
	"github.com/pyromeowww/task-scheduler/pkg/api/tasks"
	"github.com/pyromeowww/task-scheduler/tests"
)

// RunServer инициализирует и запускает HTTP-сервер.
func RunServer() {
	// Создаём логгер для записи ошибок сервера.
	logger := log.New(os.Stdout, "[server]", log.LstdFlags|log.Lshortfile)

	// Порт берём из переменной окружения TODO_PORT,
	// иначе используем порт по умолчанию.
	todoPort := os.Getenv("TODO_PORT")
	if todoPort == "" {
		todoPort = strconv.Itoa(tests.Port)
	}

	// Создаём роутер и регистрируем маршруты.
	router := http.NewServeMux()
	// Папка со статическими файлами фронтенда.
	webDir := "./web"

	router.Handle("/", http.FileServer(http.Dir(webDir)))
	router.HandleFunc("/api/nextdate", api.NextDateHandler)
	router.HandleFunc("POST /api/task", tasks.AddTaskHandler)
	router.HandleFunc("GET /api/tasks", tasks.TasksHandler)

	// Настраиваем сервер с таймаутами для защиты от зависших соединений.
	server := &http.Server{
		Addr:         ":" + todoPort,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("The server is running. Port: %s", todoPort)

	// Запускаем сервер. Ошибка возникает при остановке или сбое.
	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("Server startup error: ", err)
	}
}

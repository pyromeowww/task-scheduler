package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pyromeowww/task-scheduler/tests"
)

// RunServer инициализирует и запускает сервер
func RunServer() {
	// Создаём logger
	logger := log.New(os.Stdout, "[server]", log.LstdFlags|log.Lshortfile)

	// Пробуем получить порт из окружения, в противном случае, используем стандартное
	todoPort := os.Getenv("TODO_PORT")
	if todoPort == "" {
		todoPort = strconv.Itoa(tests.Port)
	}

	// Создаём роутер
	router := http.NewServeMux()
	// Папка с фронтенд-файлами, которые будет отдавать сервер
	webDir := "./web"

	router.Handle("/", http.FileServer(http.Dir(webDir)))

	server := &http.Server{
		Addr:         ":" + todoPort,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Сервер запущен. Порт: %s", todoPort)

	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("Ошибка запуска сервера: ", err)
	}
}

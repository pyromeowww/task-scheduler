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
	router.HandleFunc("/api/nextdate", api.NextDateHandler)
	router.HandleFunc("POST /api/task", tasks.AddTaskHandler)

	server := &http.Server{
		Addr:         ":" + todoPort,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("The server is running. Port: %s", todoPort)

	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("Server startup error: ", err)
	}
}

package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/api"
	"github.com/pyromeowww/task-scheduler/pkg/api/tasks"
	"github.com/pyromeowww/task-scheduler/pkg/settings"
)

// RunServer инициализирует и запускает HTTP-сервер.
func RunServer() {
	cfg := tasks.LoadConfig()
	// Создаём логгер для записи ошибок сервера.
	logger := log.New(os.Stdout, "[server]", log.LstdFlags|log.Lshortfile)

	// Порт берём из переменной окружения TODO_PORT,
	// иначе используем порт по умолчанию.
	port := settings.DefaultPort
	if p := os.Getenv("TODO_PORT"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 65535 {
			logger.Fatalf("invalid TODO_PORT=%q: must be integer 1..65535", p)
		}
		port = v
	}
	// Создаём роутер и регистрируем маршруты.
	router := http.NewServeMux()
	// Папка со статическими файлами фронтенда.
	webDir := "./web"

	router.Handle("/", http.FileServer(http.Dir(webDir)))

	router.HandleFunc("POST /api/signin", tasks.SigninHandler(cfg))
	// Добавление новой задачи. Ожидает JSON.
	router.Handle("POST /api/task", tasks.Auth(cfg, http.HandlerFunc(tasks.AddTaskHandler)))
	// Вычисление следующей даты для правила повторения задачи.
	router.HandleFunc("/api/nextdate", api.NextDateHandler)
	// Получение списка всех задач. Возвращает JSON-массив задач.
	router.Handle("GET /api/tasks", tasks.Auth(cfg, http.HandlerFunc(tasks.TasksHandler)))
	// Получение одной задачи по её ID.
	router.Handle("GET /api/task", tasks.Auth(cfg, http.HandlerFunc(tasks.GetTaskHandler)))
	// Обновление существующей задачи.
	router.Handle("PUT /api/task", tasks.Auth(cfg, http.HandlerFunc(tasks.UpdateTaskHandler)))
	// Выполнение существующей задачи.
	router.Handle("POST /api/task/done", tasks.Auth(cfg, http.HandlerFunc(tasks.DoneTaskHandler)))
	// Удаление существующей задачи.
	router.Handle("DELETE /api/task", tasks.Auth(cfg, http.HandlerFunc(tasks.DeleteTaskHandler)))

	// Настраиваем сервер с таймаутами для защиты от зависших соединений.
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	logger.Printf("The server is running. Port: %d", port)

	// Запускаем сервер. Ошибка возникает при остановке или сбое.
	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("Server startup error: ", err)
	}
}

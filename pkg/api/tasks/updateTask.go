package tasks

import (
	"encoding/json"
	"net/http"

	"github.com/pyromeowww/task-scheduler/pkg/db"
)

// UpdateTaskHandler обрабатывает PUT-запросы на /api/task.
// Используется для обновления существующей задачи по её ID.
func UpdateTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPut {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	// Переменная для десериализации входящих данных.
	var task db.Task

	defer req.Body.Close()

	// Декодируем JSON из тела запроса в структуру Task.
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
		return
	}

	if task.ID == "" {
		writeError(res, http.StatusBadRequest, "Task ID is required")
		return
	}

	if task.Title == "" {
		writeError(res, http.StatusBadRequest, "The task title is not specified.")
		return
	}

	// Проверяем корректность даты
	if err := checkDate(&task); err != nil {
		writeError(res, http.StatusBadRequest, err.Error())
		return
	}

	// Пытаемся обновить запись в БД
	if err := db.UpdateTask(&task); err != nil {
		writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(res, http.StatusOK, map[string]any{})
}

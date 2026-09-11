package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/api"
	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/settings"
)

// AddTaskHandler обрабатывает POST-запрос для добавления новой задачи.
func AddTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var task db.Task
	defer req.Body.Close()

	// Разбираем тело запроса в структуру задачи.
	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
		return
	}

	// Заголовок задачи обязателен.
	if task.Title == "" {
		writeError(res, http.StatusBadRequest, "The task title is not specified.")
		return
	}
	// Нормализуем и проверяем дату, а также правило повторения.
	if err := checkDate(&task); err != nil {
		writeError(res, http.StatusBadRequest, "The task title is not specified.")
		return
	}

	// Сохраняем задачу в базу данных.
	id, err := db.AddTask(&task)
	if err != nil {
		writeError(res, http.StatusInternalServerError, "Failed to save task: "+err.Error())
		return
	}

	// Возвращаем идентификатор созданной задачи.
	writeJSON(res, http.StatusOK, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// writeJSON отправляет клиенту ответ в формате JSON.
func writeJSON(res http.ResponseWriter, status int, data any) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(status)
	_ = json.NewEncoder(res).Encode(data)
}

// writeError отправляет ошибку в формате {"error": "текст"}.
func writeError(res http.ResponseWriter, status int, errText string) {
	writeJSON(res, status, map[string]string{"error": errText})
}

// checkDate нормализует дату задачи и, если она в прошлом,
// переносит её на следующую дату по правилу повторения.
func checkDate(task *db.Task) error {
	now := time.Now()
	nowStr := now.Format(settings.DateFormat)
	// Если дата не указана, используем сегодняшнюю.
	if task.Date == "" {
		task.Date = nowStr
	}

	// Проверяем, что дата соответствует формату ГГГГММДД.
	t, err := time.Parse(settings.DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format %q, expected YYYYMMDD", task.Date)
	}

	// Если дата уже прошла (раньше сегодняшнего дня), сдвигаем её вперёд.
	if api.AfterNow(now, t) {
		// Если задано правило повторения, вычисляем следующую подходящую дату,
		// иначе просто ставим сегодняшнюю.
		if len(task.Repeat) > 0 {
			next, err := api.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("incorrect repetition rule: %w", err)
			}
			task.Date = next
		} else {
			task.Date = nowStr
		}
	}
	return nil
}

package tasks

import (
	"errors"
	"net/http"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/api"
	"github.com/pyromeowww/task-scheduler/pkg/db"
)

func DoneTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем ID задачи из query-параметра.
	id := req.URL.Query().Get("id")
	if id == "" {
		writeError(res, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	// Загружаем задачу из БД по ID.
	task, err := db.GetTask(id)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrTaskNotFound):
			writeError(res, http.StatusNotFound, err.Error())
		default:
			writeError(res, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Если правило повторения не задано — задача одноразовая.
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(res, http.StatusInternalServerError, err.Error())
			return
		}
		// Успешное удаление — пустой JSON.
		writeJSON(res, http.StatusOK, map[string]any{})
		return
	}

	// Иначе вычисляем следующую дату по правилу повторения.
	nextDate, err := api.NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeError(res, http.StatusBadRequest, err.Error())
		return
	}
	// Обновляем только поле date у задачи
	if err := db.UpdateDate(nextDate, id); err != nil {
		writeError(res, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(res, http.StatusOK, map[string]any{})
}

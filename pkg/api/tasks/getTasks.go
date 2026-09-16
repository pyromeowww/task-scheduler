package tasks

import (
	"net/http"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/db"
	"github.com/pyromeowww/task-scheduler/pkg/settings"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func TasksHandler(res http.ResponseWriter, req *http.Request) {
	// Достаём query-параметр search из URL
	search := req.URL.Query().Get("search")

	var (
		tasks []*db.Task
		err   error
	)

	switch {
	case search == "":
		tasks, err = db.Tasks(settings.Limit) // в параметре максимальное количество записей
	default:
		if d, ok := tryParseDate(search); ok {
			tasks, err = db.SearchByDate(d.Format(settings.DateFormat), settings.Limit)
		} else {
			tasks, err = db.SearchByText(search, settings.Limit)
		}

		if err != nil {
			writeError(res, http.StatusInternalServerError, "Failed to get tasks: "+err.Error())
			return
		}
		// Отправляем клиенту список задач в формате JSON.
		writeJSON(res, http.StatusOK, TasksResp{Tasks: tasks})
	}
}

func GetTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := req.URL.Query().Get("id")
	if id == "" {
		writeError(res, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(res, http.StatusNotFound, "Задача не найдена")
		return
	}
	writeJSON(res, http.StatusOK, task)
}

func tryParseDate(s string) (time.Time, bool) {
	t, err := time.Parse(settings.APIDateFormat, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

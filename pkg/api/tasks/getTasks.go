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
		tasks, err = db.Tasks(50) // в параметре максимальное количество записей
	case isDate(search):
		d, _ := time.Parse("02.01.2006", search)
		tasks, err = db.SearchByDate(d.Format(settings.DateFormat), 50)
	default:
		tasks, err = db.SearchByText(search, 50)
	}
	if err != nil {
		writeError(res, http.StatusInternalServerError, "Failed to get tasks: "+err.Error())
		return
	}
	// Отправляем клиенту список задач в формате JSON.
	writeJSON(res, http.StatusOK, TasksResp{Tasks: tasks})
}

func GetTaskHandler(res http.ResponseWriter, req *http.Request) {
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

func isDate(s string) bool {
	_, err := time.Parse("02.01.2006", s)
	return err == nil
}

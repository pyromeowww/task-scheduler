package tasks

import (
	"net/http"

	"github.com/pyromeowww/task-scheduler/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func TasksHandler(res http.ResponseWriter, req *http.Request) {
	tasks, err := db.Tasks(50) // в параметре максимальное количество записей
	if err != nil {
		// возвращает ошибку в JSON
		writeError(res, http.StatusInternalServerError, "Failed to get tasks: "+err.Error())
		return
	}

	// Отправляем клиенту список задач в формате JSON.
	writeJSON(res, http.StatusOK, TasksResp{
		Tasks: tasks,
	})
}

package tasks

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pyromeowww/task-scheduler/pkg/db"
)

const DateFormat = "20060102"

func AddTaskHandler(res http.ResponseWriter, req *http.Request) {
	var task db.Task

	if err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
		return
	}

	if task.Title == "" {
		writeError(res, http.StatusBadRequest, "The task title is not specified.")
		return
	}
}

// writeJSON отправляет структурированный ответ в JSON
func writeJSON(res http.ResponseWriter, status int, data any) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(status)
	_ = json.NewEncoder(res).Encode(data)
}

// writeError отправляет ошибку в формате {"error": "текст"}
func writeError(res http.ResponseWriter, status int, errText string) {
	writeJSON(res, status, map[string]string{"error": errText})
}

func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	t, err := time.Parse(DateFormat, task.Date)
}

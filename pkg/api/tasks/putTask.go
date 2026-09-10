package tasks

import (
	"encoding/json"
	"net/http"

	"github.com/pyromeowww/task-scheduler/pkg/db"
)

func UpdateTaskHadler(res http.ResponseWriter, req *http.Request) {
	var task db.Task

	err := json.NewDecoder(req.Body).Decode(&task); err != nil {
		writeError(res, http.StatusBadRequest, "JSON deserialization error: "+err.Error())
		return
	}
}
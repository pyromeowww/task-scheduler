package tasks

import (
	"errors"
	"net/http"

	"github.com/pyromeowww/task-scheduler/pkg/db"
)

func DeleteTaskHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		writeError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := req.URL.Query().Get("id")
	if id == "" {
		writeError(res, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	if err := db.DeleteTask(id); err != nil {
		switch {
		case errors.Is(err, db.ErrTaskNotFound):
			writeError(res, http.StatusNotFound, err.Error())
		default:
			writeError(res, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(res, http.StatusOK, map[string]any{})
}

package api

import (
	"FINAL-13-14/pkg/db"
	"net/http"
)

const LimitGetTask = 50

// TasksResponse структура ответа
type TasksResponse struct {
	Tasks []db.Task `json:"tasks"` // используем Task из db
}

// TasksHandler обрабатывает GET /api/tasks
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметр search из URL
	search := r.URL.Query().Get("search")

	var tasks []db.Task
	var err error

	if search != "" {
		tasks, err = db.GetTasks(LimitGetTask, search)
	} else {
		tasks, err = db.GetTasks(LimitGetTask)
	}

	if err != nil {
		sendError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	// Убеждаемся, что tasks не nil
	if tasks == nil {
		tasks = make([]db.Task, 0)
	}

	writeJSON(w, TasksResponse{
		Tasks: tasks,
	})
}

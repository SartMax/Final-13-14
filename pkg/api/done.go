package api

import (
	"encoding/json"
	"net/http"
	"time"

	"FINAL-13-14/pkg/db"
)

// DoneHandler обрабатывает POST /api/task/done
func DoneHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из query параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, http.StatusBadRequest, "Не указан идентификатор")
		return
	}

	// Получаем задачу из БД
	task, err := db.GetTask(id)
	if err != nil {
		if err.Error() == "task not found" {
			sendError(w, http.StatusNotFound, "Задача не найдена")
		} else {
			sendError(w, http.StatusInternalServerError, "Ошибка базы данных: "+err.Error())
		}
		return
	}

	// Если нет правила повторения - удаляем задачу
	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			if err.Error() == "task not found" {
				sendError(w, http.StatusNotFound, "Задача не найдена")
			} else {
				sendError(w, http.StatusInternalServerError, "Ошибка при удалении задачи: "+err.Error())
			}
			return
		}
	} else {
		// Есть правило повторения - вычисляем следующую дату
		now := time.Now()

		// Вычисляем следующую дату
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Ошибка при вычислении следующей даты: "+err.Error())
			return
		}

		// Обновляем только дату
		err = db.UpdateTaskDate(id, nextDate)
		if err != nil {
			if err.Error() == "task not found" {
				sendError(w, http.StatusNotFound, "Задача не найдена")
			} else {
				sendError(w, http.StatusInternalServerError, "Ошибка при обновлении даты: "+err.Error())
			}
			return
		}
	}

	// Возвращаем пустой JSON при успехе
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

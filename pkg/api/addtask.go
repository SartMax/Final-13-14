package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"FINAL-13-14/pkg/db"
)

// AddTaskHandler обрабатывает POST /api/task для добавления задачи,
// GET /api/task?id= для получения задачи,
// и PUT /api/task для обновления задачи
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostTask(w, r)
	case http.MethodPut:
		handlePutTask(w, r)
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodDelete:
		handleDeleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetTask обрабатывает GET /api/task?id=123
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из query параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "Не указан идентификатор")
		return
	}

	// Получаем задачу из БД
	task, err := db.GetTask(id)
	if err != nil {
		if err.Error() == "task not found" {
			sendError(w, "Задача не найдена")
		} else {
			sendError(w, "Ошибка базы данных: "+err.Error())
		}
		return
	}

	// Отправляем задачу
	writeJSON(w, task)
}

// handlePostTask обрабатывает POST /api/task (добавление)
func handlePostTask(w http.ResponseWriter, r *http.Request) {
	// Декодируем JSON в структуру Task из пакета db
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "Invalid JSON format")
		return
	}

	// Валидируем и обрабатываем задачу
	if err := processTask(&task); err != nil {
		sendError(w, err.Error())
		return
	}

	// Сохраняем в базу данных, используя функцию AddTask из пакета db
	id, err := db.AddTask(&task)
	if err != nil {
		sendError(w, "Database error: "+err.Error())
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

// handlePutTask обрабатывает PUT /api/task
func handlePutTask(w http.ResponseWriter, r *http.Request) {
	// Декодируем JSON в структуру Task
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendError(w, "Invalid JSON format")
		return
	}

	// Проверяем, что указан ID
	if task.ID == "" {
		sendError(w, "Не указан идентификатор задачи")
		return
	}

	// Валидируем и обрабатываем задачу
	if err := processTask(&task); err != nil {
		sendError(w, err.Error())
		return
	}

	// Обновляем задачу в базе данных
	err := db.UpdateTask(&task)
	if err != nil {
		if err.Error() == "task not found" {
			sendError(w, "Задача не найдена")
		} else {
			sendError(w, "Database error: "+err.Error())
		}
		return
	}

	// Возвращаем пустой JSON при успехе
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// processTask проверяет и корректирует поля задачи
func processTask(task *db.Task) error {
	// Проверяем обязательное поле title
	if task.Title == "" {
		return errors.New("title is required")
	}

	now := time.Now()
	nowDate := now.Format(DateFormat)

	// Если дата не указана, используем сегодняшнюю
	if task.Date == "" {
		task.Date = nowDate
	}

	// Проверяем формат даты
	taskTime, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return errors.New("invalid date format, expected YYYYMMDD")
	}

	// Если есть правило повторения, проверяем его корректность
	var nextDate string
	if task.Repeat != "" {
		nextDate, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("invalid repeat rule: " + err.Error())
		}
	}

	// Проверяем, что дата не в прошлом
	if isAfter(now, taskTime) {
		if task.Repeat == "" {
			// Если правила нет, ставим сегодняшнюю дату
			task.Date = nowDate
		} else {
			// Если есть правило, используем следующую дату
			task.Date = nextDate
		}
	}

	return nil
}

// isAfter возвращает true, если now > date
func isAfter(now, date time.Time) bool {
	y1, m1, d1 := now.Date()
	y2, m2, d2 := date.Date()

	nowDate := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	dateDate := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)

	return nowDate.After(dateDate)
}
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из query параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "Не указан идентификатор")
		return
	}

	// Удаляем задачу из БД
	err := db.DeleteTask(id)
	if err != nil {
		if err.Error() == "task not found" {
			sendError(w, "Задача не найдена")
		} else {
			sendError(w, "Ошибка базы данных: "+err.Error())
		}
		return
	}

	// Возвращаем пустой JSON при успехе
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

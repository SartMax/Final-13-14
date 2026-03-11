package db

import (
	"database/sql"
	"fmt"
	"time"
)

const DateFormat = "20060102"

type Task struct {
	ID      string `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

// AddTask добавляет новую задачу в базу данных
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	result, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetTasks(limit int, search ...string) ([]Task, error) {
	var rows *sql.Rows
	var err error

	// Если есть параметр поиска
	if len(search) > 0 && search[0] != "" {
		searchTerm := search[0]

		// Проверяем, является ли поиск датой в формате 02.01.2006
		if t, err := time.Parse("02.01.2006", searchTerm); err == nil {
			// Поиск по конкретной дате
			dbDate := t.Format(DateFormat)
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`
			rows, err = DB.Query(query, dbDate, limit)
		} else {
			// Поиск по тексту (без учета регистра через LIKE)
			query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`
			pattern := "%" + searchTerm + "%"
			rows, err = DB.Query(query, pattern, pattern, limit)
		}
	} else {
		// Без поиска
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
		rows, err = DB.Query(query, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Возвращаем пустой слайс вместо nil
	if tasks == nil {
		tasks = make([]Task, 0)
	}

	return tasks, nil
}
func GetTask(id string) (*Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	var task Task
	err := DB.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	return &task, nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	result, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask удаляет задачу по ID
func DeleteTask(id string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `DELETE FROM scheduler WHERE id = ?`

	result, err := DB.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// UpdateTaskDate обновляет только дату задачи
func UpdateTaskDate(id, newDate string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`

	result, err := DB.Exec(query, newDate, id)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

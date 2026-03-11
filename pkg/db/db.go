package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// SQL для создания таблицы и индекса
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL,
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);

CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
`

var DB *sql.DB

// Init инициализирует подключение к базе данных
func Init(dbFile string) error {
	// Проверяем существование файла БД
	var install bool
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
		log.Printf("Файл базы данных %s не найден, будет создан новый", dbFile)
	}

	// Открываем базу данных
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	// Проверяем соединение
	if err = db.Ping(); err != nil {
		db.Close()
		return err
	}

	// Если файл не существовал, создаем таблицу
	if install {
		log.Println("Создание таблицы scheduler и индекса...")
		_, err = db.Exec(schema)
		if err != nil {
			db.Close()
			return err
		}
		log.Println("Таблица успешно создана")
	}

	DB = db
	return nil
}

// Close закрывает соединение с базой данных
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

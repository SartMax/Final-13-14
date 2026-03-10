package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	Port   int
	WebDir string
	DBFile string
}

func Load() *Config {
	port := 7540 // порт по умолчанию

	// Проверяем переменную окружения TODO_PORT
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	// Получаем путь к директории проекта (где находится go.mod)
	projectDir := getProjectDir()

	// Путь к файлу БД
	dbFile := filepath.Join(projectDir, "scheduler.db")
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		dbFile = envDBFile
	}

	// Путь к web директории
	webDir := filepath.Join(projectDir, "web")

	return &Config{
		Port:   port,
		WebDir: webDir,
		DBFile: dbFile,
	}
}

// getProjectDir возвращает путь к директории проекта (где находится go.mod)
func getProjectDir() string {
	// Получаем путь к текущему файлу (config.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		// Если не можем получить путь, используем текущую директорию
		wd, _ := os.Getwd()
		return wd
	}

	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return projectDir
}

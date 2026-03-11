package main

import (
	"log"

	"FINAL-13-14/pkg/config"
	"FINAL-13-14/pkg/server"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	log.Printf("Загружена конфигурация:")
	log.Printf("  Порт: %d", cfg.Port)
	log.Printf("  Web директория: %s", cfg.WebDir)
	log.Printf("  Файл БД: %s", cfg.DBFile)

	// Создаем сервер
	srv := server.New(&server.Config{
		Port:   cfg.Port,
		WebDir: cfg.WebDir,
		DBFile: cfg.DBFile,
	})

	// Запускаем сервер
	if err := srv.Start(); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}

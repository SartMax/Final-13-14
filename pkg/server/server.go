package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"FINAL-13-14/pkg/api"
	"FINAL-13-14/pkg/db"
)

type Server struct {
	config *Config
	router *chi.Mux
}

type Config struct {
	Port   int
	WebDir string
	DBFile string
}

func New(cfg *Config) *Server {
	return &Server{
		config: cfg,
		router: chi.NewRouter(),
	}
}

func (s *Server) Start() error {
	// Инициализируем базу данных
	if err := db.Init(s.config.DBFile); err != nil {
		return fmt.Errorf("ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	log.Printf("База данных подключена: %s", s.config.DBFile)

	// Проверяем существование директории web
	if _, err := os.Stat(s.config.WebDir); os.IsNotExist(err) {
		return fmt.Errorf("директория web не найдена по пути: %s", s.config.WebDir)
	}

	log.Printf("Обслуживание файлов из: %s", s.config.WebDir)

	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// API маршруты
	s.router.Get("/api/nextdate", api.HandleNextDate)
	s.router.HandleFunc("/api/task", api.AddTaskHandler)
	s.router.Get("/api/tasks", api.TasksHandler)
	s.router.Post("/api/task/done", api.DoneHandler)
	// Статические файлы
	fileServer := http.FileServer(http.Dir(s.config.WebDir))
	s.router.Handle("/*", fileServer)

	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Сервер запущен на порту %d", s.config.Port)
	log.Printf("Откройте http://localhost:%d/ в браузере", s.config.Port)

	return http.ListenAndServe(addr, s.router)
}

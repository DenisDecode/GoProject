// Package app собирает приложение: конфигурация -> БД -> миграции ->
// репозитории -> шаблоны -> обработчики -> HTTP-сервер.
package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expense-tracker/internal/config"
	"expense-tracker/internal/database"
	"expense-tracker/internal/handlers"
	"expense-tracker/internal/render"
	"expense-tracker/internal/repository"
)

func Run() {
	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("не удалось открыть базу данных: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db, cfg.MigrationsDir); err != nil {
		log.Fatalf("ошибка миграций: %v", err)
	}

	renderer, err := render.New(cfg.TemplatesDir)
	if err != nil {
		log.Fatalf("ошибка загрузки шаблонов: %v", err)
	}

	expenses := repository.NewExpenseRepository(db)
	categories := repository.NewCategoryRepository(db)
	users := repository.NewUserRepository(db)
	sessions := repository.NewSessionRepository(db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go cleanupSessions(ctx, sessions)

	h := handlers.New(cfg, renderer, expenses, categories, users, sessions)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("сервер запущен: http://localhost:%s (БД: %s)", cfg.Port, cfg.DBPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ошибка сервера: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("остановка сервера...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("ошибка при остановке: %v", err)
	}
	log.Println("сервер остановлен")
}

// cleanupSessions периодически удаляет просроченные сессии из БД.
func cleanupSessions(ctx context.Context, sessions *repository.SessionRepository) {
	clean := func() {
		n, err := sessions.DeleteExpired(context.Background())
		if err != nil {
			log.Printf("очистка сессий: %v", err)
			return
		}
		if n > 0 {
			log.Printf("удалено просроченных сессий: %d", n)
		}
	}
	clean()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		}
	}
}

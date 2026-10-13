package main

import (
	"html/template"
	"log"
	"net/http"

	"GoProject/internal/db"
	"GoProject/internal/handlers"
	"GoProject/internal/repository"
)

func main() {
	// 1. Подключаем базу данных SQLite
	database, err := db.NewDB("expenses.db")
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer database.Close()

	// 2. Инициализируем репозиторий
	repo := repository.NewRepository(database)

	// 3. Загружаем HTML-шаблоны
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Ошибка загрузки шаблонов: %v", err)
	}

	// 4. Создаем экземпляр хэндлеров, передавая репозиторий и шаблоны
	handler := handlers.NewHandler(repo, tmpl)

	// 5. Настраиваем маршрутизатор (роутер)
	mux := http.NewServeMux()

	// Статические файлы
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Основные маршруты приложения
	mux.HandleFunc("GET /expenses", handler.ExpensesHandler)
	mux.HandleFunc("POST /expenses/add", handler.AddExpenseHandler)

	// Точный редирект с корня с помощью синтаксиса {$} (Go 1.22+)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
	})

	// 6. Запуск сервера
	log.Println("Сервер успешно запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}

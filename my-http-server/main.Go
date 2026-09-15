package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Создаем новый маршрутизатор (ServeMux)
	mux := http.NewServeMux()

	// 1. Главная страница (текст-приветствие)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Строгая проверка пути, чтобы корень не перехватывал другие несуществующие страницы
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, "Привет! Добро пожаловать на наш первый веб-сервер на Go.")
	})

	// 2. Страница /about (описание проекта)
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Это страница с описанием учебного проекта на чистом пакете net/http.")
	})

	// 3. Страница /ping с ручной проверкой HTTP-метода (Задание 3)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		// Если метод отличный от GET, возвращаем статус 405 Method Not Allowed
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, "405 Method Not Allowed. Используйте только GET.")
			return
		}

		// Если метод GET — возвращаем 200 OK и слово "pong"
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	})

	// Запуск HTTP-сервера на порту 8080
	fmt.Println("Сервер успешно запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v\n", err)
	}
}

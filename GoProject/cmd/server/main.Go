package main

import (
	"fmt"
	"net/http"
	"my-server/internal/handlers"
)

func main() {
	mux := http.NewServeMux()


	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))


	mux.HandleFunc("/", handlers.ExpenseHandler)


	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Это страница с описанием учебного проекта на чистом пакете net/http.")
	})


	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, "405 Method Not Allowed. Используйте только GET.")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	})


	fmt.Println("Сервер успешно запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v\n", err)
	}
}

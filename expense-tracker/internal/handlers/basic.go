package handlers

import "net/http"

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, http.StatusOK, "home.html", h.newTemplateData(r))
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, http.StatusOK, "about.html", h.newTemplateData(r))
}

// Ping: GET -> 200 "pong", любой другой метод -> 405 + заголовок Allow.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// Panic — искусственная паника для демонстрации middleware Recover.
func (h *Handler) Panic(w http.ResponseWriter, r *http.Request) {
	panic("искусственная паника для проверки middleware Recover")
}

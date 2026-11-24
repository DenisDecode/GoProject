package handlers

import (
	"net/http"

	"expense-tracker/internal/middleware"
)

// Routes регистрирует маршруты (шаблоны Go 1.22+: "МЕТОД /путь/{параметр}")
// и оборачивает их в цепочку middleware.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Статика
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(h.cfg.StaticDir))))

	// Этап 1
	mux.HandleFunc("GET /{$}", h.Home) // {$} — точное совпадение с "/"
	mux.HandleFunc("GET /about", h.About)
	mux.HandleFunc("/ping", h.Ping) // метод проверяется вручную внутри -> 405

	// Этап 5: публичные страницы авторизации
	mux.HandleFunc("GET /register", h.RegisterPage)
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)

	// Защищённые маршруты
	protected := func(f http.HandlerFunc) http.Handler { return middleware.RequireAuth(f) }

	mux.Handle("GET /expenses", protected(h.ExpenseList))
	mux.Handle("GET /expenses/new", protected(h.ExpenseNew))
	mux.Handle("POST /expenses", protected(h.ExpenseCreate))
	mux.Handle("GET /expenses/{id}/edit", protected(h.ExpenseEdit))
	mux.Handle("POST /expenses/{id}/edit", protected(h.ExpenseUpdate))
	mux.Handle("POST /expenses/{id}/delete", protected(h.ExpenseDelete))

	mux.Handle("GET /stats", protected(h.StatsPage))
	mux.Handle("GET /api/stats/by-category", protected(h.APIStatsByCategory))
	mux.Handle("GET /api/stats/by-month", protected(h.APIStatsByMonth))

	// Этап 7: маршрут для проверки recover (отключается ENABLE_DEBUG_ROUTES=false)
	if h.cfg.EnableDebugRoutes {
		mux.HandleFunc("GET /debug/panic", h.Panic)
	}

	// Всё остальное — кастомная 404
	mux.HandleFunc("/", h.NotFound)

	// Порядок: Logging -> Recover -> LoadUser -> mux -> (RequireAuth) -> обработчик
	return middleware.Chain(mux,
		middleware.Logging,
		middleware.Recover(h.PanicResponse),
		middleware.LoadUser(h.sessions, h.users, h.cfg.CookieSecure),
	)
}

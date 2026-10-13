package handlers

import (
	"GoProject/internal/repository"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	repo *repository.Repository
	tmpl *template.Template
}

func NewHandler(repo *repository.Repository, tmpl *template.Template) *Handler {
	return &Handler{repo: repo, tmpl: tmpl}
}

func (h *Handler) ExpensesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Читаем параметры фильтрации из URL (Этап 4)
	categoryID := r.URL.Query().Get("category")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")

	// 2. Получаем данные из базы через репозиторий (вместо памяти)
	expenses, err := h.repo.GetFilteredExpenses(categoryID, dateFrom, dateTo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories, err := h.repo.GetAllCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Передаем данные в шаблон
	data := map[string]interface{}{
		"Expenses":   expenses,
		"Categories": categories,
		"SelCat":     categoryID,
		"DateFrom":   dateFrom,
		"DateTo":     dateTo,
	}

	// ИЗМЕНЕНИЕ: проверяем ошибку выполнения шаблона
	err = h.tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Ошибка отрисовки шаблона: %v", err)
		http.Error(w, "Ошибка сервера при отрисовке страницы", http.StatusInternalServerError)
		return
	}
}

// AddExpenseHandler обрабатывает POST-запрос от формы добавления расхода
func (h *Handler) AddExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// 1. Парсим данные формы из HTML
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
		return
	}

	// 2. Получаем значения по их name из index.html
	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	date := r.FormValue("date")
	categoryIDStr := r.FormValue("category_id")

	// 3. Конвертируем строку суммы в float64
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	// 4. Конвертируем ID категории в int
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		http.Error(w, "Некорректная категория", http.StatusBadRequest)
		return
	}

	// 5. Вызываем готовый метод репозитория для сохранения в БД
	err = h.repo.CreateExpense(amount, description, date, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Перенаправляем пользователя обратно на страницу со списком
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

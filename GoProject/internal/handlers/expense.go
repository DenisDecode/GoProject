package handlers

import (
	"html/template"
	"my-server/internal/models"
	"net/http"
	"strconv"
	"time"
)

// Хранилище в памяти (взамен базы данных на этом этапе)
var Expenses []models.Expense

// Инициализируем несколькими тестовыми записями для наглядности
func init() {
	Expenses = []models.Expense{
		{Amount: 1500.0, Description: "Продукты", Date: time.Now().AddDate(0, 0, -1)},
		{Amount: 450.0, Description: "Кофе и десерт", Date: time.Now()},
	}
}

// ExpenseHandler управляет отображением и добавлением трат
func ExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// 1. Разбираем данные формы
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
			return
		}

		amountStr := r.FormValue("amount")
		description := r.FormValue("description")
		dateStr := r.FormValue("date")

		// 2. Валидация и конвертация данных
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "Некорректная сумма", http.StatusBadRequest)
			return
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			date = time.Now() // если дата не указана, берем текущую
		}

		// 3. Сохраняем в наш срез в памяти
		newExpense := models.Expense{
			Amount:      amount,
			Description: description,
			Date:        date,
		}
		Expenses = append(Expenses, newExpense)

		// 4. Паттерн Post/Redirect/Get (PRG) — защищает от повторной отправки формы
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Если метод GET — рендерим страницу со списком трат
	// Парсим базовый шаблон (layout) и контентную страницу
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/index.html")
	if err != nil {
		http.Error(w, "Ошибка компиляции шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем срез трат в шаблон
	err = tmpl.ExecuteTemplate(w, "layout", Expenses)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
	}
}

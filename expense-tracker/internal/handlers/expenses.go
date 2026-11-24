package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"expense-tracker/internal/forms"
	"expense-tracker/internal/models"
)

// ExpenseList — список трат текущего пользователя с фильтрами из query-параметров.
func (h *Handler) ExpenseList(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	q := r.URL.Query()
	filter, filterErrs := forms.ParseExpenseFilter(q)

	expenses, err := h.expenses.GetAll(r.Context(), user.ID, filter)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	total, count, err := h.expenses.Summary(r.Context(), user.ID, filter)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	categories, err := h.categories.GetAll(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	data := h.newTemplateData(r)
	data.Expenses = expenses
	data.Total = total
	data.Count = count
	data.Categories = categories
	data.FilterErrors = filterErrs
	data.Filter = filterView{
		Category: q.Get("category"),
		From:     q.Get("from"),
		To:       q.Get("to"),
		Active:   q.Get("category") != "" || q.Get("from") != "" || q.Get("to") != "",
	}
	h.render(w, r, http.StatusOK, "expenses.html", data)
}

func (h *Handler) ExpenseNew(w http.ResponseWriter, r *http.Request) {
	form := forms.NewExpenseForm()
	form.Date = time.Now().Format(forms.DateLayout)
	h.renderExpenseForm(w, r, http.StatusOK, form, nil)
}

// ExpenseCreate: валидация -> при ошибке повторный рендер формы (422),
// при успехе — сохранение и redirect (Post/Redirect/Get).
func (h *Handler) ExpenseCreate(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest)
		return
	}

	form := forms.ExpenseFormFromRequest(r)
	expense, ok, err := h.validateExpense(r.Context(), form)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	if !ok {
		h.renderExpenseForm(w, r, http.StatusUnprocessableEntity, form, nil)
		return
	}

	expense.UserID = user.ID
	if _, err := h.expenses.Create(r.Context(), expense); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) ExpenseEdit(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	id, ok := parseID(r)
	if !ok {
		h.NotFound(w, r)
		return
	}
	expense, err := h.expenses.GetByID(r.Context(), user.ID, id)
	if errors.Is(err, models.ErrNotFound) {
		h.NotFound(w, r)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}
	h.renderExpenseForm(w, r, http.StatusOK, forms.ExpenseFormFromExpense(expense), &expense)
}

func (h *Handler) ExpenseUpdate(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	id, ok := parseID(r)
	if !ok {
		h.NotFound(w, r)
		return
	}
	existing, err := h.expenses.GetByID(r.Context(), user.ID, id)
	if errors.Is(err, models.ErrNotFound) {
		h.NotFound(w, r)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest)
		return
	}

	form := forms.ExpenseFormFromRequest(r)
	expense, valid, err := h.validateExpense(r.Context(), form)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	if !valid {
		h.renderExpenseForm(w, r, http.StatusUnprocessableEntity, form, &existing)
		return
	}

	expense.ID = id
	expense.UserID = user.ID
	if err := h.expenses.Update(r.Context(), expense); errors.Is(err, models.ErrNotFound) {
		h.NotFound(w, r)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) ExpenseDelete(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	id, ok := parseID(r)
	if !ok {
		h.NotFound(w, r)
		return
	}
	if err := h.expenses.Delete(r.Context(), user.ID, id); errors.Is(err, models.ErrNotFound) {
		h.NotFound(w, r)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// validateExpense: проверки формы + проверка существования категории в БД.
func (h *Handler) validateExpense(ctx context.Context, form *forms.ExpenseForm) (models.Expense, bool, error) {
	expense := form.Validate(time.Now())
	if expense.CategoryID > 0 {
		exists, err := h.categories.Exists(ctx, expense.CategoryID)
		if err != nil {
			return expense, false, err
		}
		if !exists {
			form.AddError("category", "Такой категории не существует")
		}
	}
	return expense, form.Valid(), nil
}

func (h *Handler) renderExpenseForm(w http.ResponseWriter, r *http.Request, status int, form *forms.ExpenseForm, expense *models.Expense) {
	categories, err := h.categories.GetAll(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	data := h.newTemplateData(r)
	data.Form = form
	data.Categories = categories
	data.Expense = expense
	h.render(w, r, status, "expense_form.html", data)
}

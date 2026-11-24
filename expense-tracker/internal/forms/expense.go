package forms

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"expense-tracker/internal/models"
)

const (
	MaxAmount         = 1_000_000_000
	MaxDescriptionLen = 200
)

var (
	amountRe = regexp.MustCompile(`^\d{1,12}(\.\d{1,2})?$`)
	minDate  = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

// ExpenseForm хранит СЫРЫЕ строки из формы, чтобы при ошибке
// показать пользователю ровно то, что он ввёл.
type ExpenseForm struct {
	Base
	Amount      string
	Description string
	Date        string
	CategoryID  string
}

func NewExpenseForm() *ExpenseForm {
	return &ExpenseForm{Base: Base{Errors: map[string]string{}}}
}

func ExpenseFormFromRequest(r *http.Request) *ExpenseForm {
	f := NewExpenseForm()
	f.Amount = strings.TrimSpace(r.PostFormValue("amount"))
	f.Description = strings.TrimSpace(r.PostFormValue("description"))
	f.Date = strings.TrimSpace(r.PostFormValue("date"))
	f.CategoryID = strings.TrimSpace(r.PostFormValue("category_id"))
	return f
}

func ExpenseFormFromExpense(e models.Expense) *ExpenseForm {
	f := NewExpenseForm()
	f.Amount = strconv.FormatFloat(e.Amount, 'f', 2, 64)
	f.Description = e.Description
	f.Date = e.Date
	if e.CategoryID > 0 {
		f.CategoryID = strconv.FormatInt(e.CategoryID, 10)
	}
	return f
}

// Validate проверяет все поля и возвращает разобранную трату.
// Если есть ошибки — они записываются в f.Errors (см. f.Valid()).
func (f *ExpenseForm) Validate(now time.Time) models.Expense {
	var e models.Expense

	// Сумма: положительное число, не более 2 знаков после запятой.
	amount := strings.NewReplacer(" ", "", "\u00a0", "", ",", ".").Replace(f.Amount)
	switch {
	case amount == "":
		f.AddError("amount", "Укажите сумму")
	case strings.HasPrefix(amount, "-"):
		f.AddError("amount", "Сумма должна быть больше нуля")
	case !amountRe.MatchString(amount):
		f.AddError("amount", "Сумма должна быть числом, не более 2 знаков после запятой (например, 250 или 199,90)")
	default:
		v, err := strconv.ParseFloat(amount, 64)
		switch {
		case err != nil:
			f.AddError("amount", "Сумма должна быть числом")
		case v <= 0:
			f.AddError("amount", "Сумма должна быть больше нуля")
		case v > MaxAmount:
			f.AddError("amount", fmt.Sprintf("Сумма не может превышать %d", MaxAmount))
		default:
			e.Amount = v
		}
	}

	// Описание: обязательно, ограничение длины.
	switch {
	case f.Description == "":
		f.AddError("description", "Введите описание")
	case utf8.RuneCountInString(f.Description) > MaxDescriptionLen:
		f.AddError("description", fmt.Sprintf("Описание не длиннее %d символов", MaxDescriptionLen))
	default:
		e.Description = f.Description
	}

	// Дата: формат ГГГГ-ММ-ДД, не раньше 2000 года и не в будущем.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if f.Date == "" {
		f.AddError("date", "Укажите дату")
	} else if d, err := time.Parse(DateLayout, f.Date); err != nil {
		f.AddError("date", "Дата должна быть в формате ГГГГ-ММ-ДД")
	} else if d.Before(minDate) {
		f.AddError("date", "Дата не может быть раньше 01.01.2000")
	} else if d.After(today) {
		f.AddError("date", "Дата не может быть в будущем")
	} else {
		e.Date = f.Date
	}

	// Категория: обязательна, положительное целое (существование проверяет обработчик через репозиторий).
	if f.CategoryID == "" {
		f.AddError("category", "Выберите категорию")
	} else if id, err := strconv.ParseInt(f.CategoryID, 10, 64); err != nil || id <= 0 {
		f.AddError("category", "Некорректная категория")
	} else {
		e.CategoryID = id
	}

	return e
}

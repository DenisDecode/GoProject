package forms

import (
	"net/url"
	"strconv"
	"strings"

	"expense-tracker/internal/models"
)

// ParseExpenseFilter разбирает query-параметры ?category=&from=&to=.
// Некорректные значения не ломают страницу: фильтр игнорируется, а
// пользователь получает понятное сообщение.
func ParseExpenseFilter(q url.Values) (models.ExpenseFilter, []string) {
	var f models.ExpenseFilter
	var errs []string

	if v := strings.TrimSpace(q.Get("category")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			errs = append(errs, "Некорректная категория — фильтр по категории не применён")
		} else {
			f.CategoryID = id
		}
	}

	if v := strings.TrimSpace(q.Get("from")); v != "" {
		if IsValidDate(v) {
			f.From = v
		} else {
			errs = append(errs, "Дата «с» должна быть в формате ГГГГ-ММ-ДД — фильтр не применён")
		}
	}
	if v := strings.TrimSpace(q.Get("to")); v != "" {
		if IsValidDate(v) {
			f.To = v
		} else {
			errs = append(errs, "Дата «по» должна быть в формате ГГГГ-ММ-ДД — фильтр не применён")
		}
	}

	if f.From != "" && f.To != "" && f.From > f.To {
		errs = append(errs, "Дата «с» позже даты «по» — фильтр по датам не применён")
		f.From, f.To = "", ""
	}
	return f, errs
}

// Package repository инкапсулирует весь SQL-код приложения.
// Все запросы используют плейсхолдеры ? (защита от SQL-инъекций).
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"expense-tracker/internal/models"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// LEFT JOIN — название категории получаем одним запросом (без N+1).
const expenseSelect = `
SELECT e.id, e.user_id, e.amount, e.description, e.date,
       e.category_id, COALESCE(c.name, 'Без категории')
FROM expenses e
LEFT JOIN categories c ON c.id = e.category_id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanExpense(s rowScanner) (models.Expense, error) {
	var e models.Expense
	var categoryID sql.NullInt64
	err := s.Scan(&e.ID, &e.UserID, &e.Amount, &e.Description, &e.Date, &categoryID, &e.CategoryName)
	if err != nil {
		return e, err
	}
	if categoryID.Valid {
		e.CategoryID = categoryID.Int64
	}
	return e, nil
}

func nullableID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

// buildExpenseWhere динамически собирает WHERE: условие по пользователю есть всегда,
// остальные добавляются только если фильтр задан. Значения идут только в args.
func buildExpenseWhere(userID int64, f models.ExpenseFilter) (string, []any) {
	conds := []string{"e.user_id = ?"}
	args := []any{userID}

	if f.CategoryID > 0 {
		conds = append(conds, "e.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if f.From != "" {
		conds = append(conds, "e.date >= ?")
		args = append(args, f.From)
	}
	if f.To != "" {
		conds = append(conds, "e.date <= ?")
		args = append(args, f.To)
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

func (r *ExpenseRepository) Create(ctx context.Context, e models.Expense) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO expenses (user_id, category_id, amount, description, date) VALUES (?, ?, ?, ?, ?)`,
		e.UserID, nullableID(e.CategoryID), e.Amount, e.Description, e.Date)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAll возвращает траты пользователя с учётом фильтров.
func (r *ExpenseRepository) GetAll(ctx context.Context, userID int64, f models.ExpenseFilter) ([]models.Expense, error) {
	where, args := buildExpenseWhere(userID, f)
	query := expenseSelect + where + " ORDER BY e.date DESC, e.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]models.Expense, 0)
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

// Summary — сумма и количество трат по тем же фильтрам (агрегация в БД).
func (r *ExpenseRepository) Summary(ctx context.Context, userID int64, f models.ExpenseFilter) (float64, int, error) {
	where, args := buildExpenseWhere(userID, f)
	var total float64
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(e.amount), 0.0), COUNT(*) FROM expenses e`+where, args...).
		Scan(&total, &count)
	return total, count, err
}

func (r *ExpenseRepository) GetByID(ctx context.Context, userID, id int64) (models.Expense, error) {
	row := r.db.QueryRowContext(ctx, expenseSelect+" WHERE e.id = ? AND e.user_id = ?", id, userID)
	e, err := scanExpense(row)
	if errors.Is(err, sql.ErrNoRows) {
		return e, models.ErrNotFound
	}
	return e, err
}

func (r *ExpenseRepository) Update(ctx context.Context, e models.Expense) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.Date, nullableID(e.CategoryID), e.ID, e.UserID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (r *ExpenseRepository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}

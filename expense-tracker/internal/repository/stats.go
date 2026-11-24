package repository

import (
	"context"
	"fmt"

	"expense-tracker/internal/models"
)

// SumByCategory — сумма трат по категориям за период (SUM + GROUP BY в БД).
func (r *ExpenseRepository) SumByCategory(ctx context.Context, userID int64, from, to string) ([]models.CategoryStat, error) {
	where, args := buildExpenseWhere(userID, models.ExpenseFilter{From: from, To: to})
	query := `
SELECT COALESCE(c.name, 'Без категории') AS category,
       SUM(e.amount) AS total,
       COUNT(*)      AS cnt
FROM expenses e
LEFT JOIN categories c ON c.id = e.category_id` + where + `
GROUP BY e.category_id, c.name
ORDER BY total DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]models.CategoryStat, 0)
	for rows.Next() {
		var s models.CategoryStat
		if err := rows.Scan(&s.Category, &s.Total, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// SumByMonth — сумма трат по месяцам за год. Всегда возвращает 12 элементов
// (месяцы без трат — с нулями), чтобы график был полным.
func (r *ExpenseRepository) SumByMonth(ctx context.Context, userID int64, year int) ([]models.MonthStat, error) {
	query := `
SELECT CAST(substr(e.date, 6, 2) AS INTEGER) AS month,
       SUM(e.amount),
       COUNT(*)
FROM expenses e
WHERE e.user_id = ? AND e.date >= ? AND e.date <= ?
GROUP BY month
ORDER BY month`

	rows, err := r.db.QueryContext(ctx, query, userID,
		fmt.Sprintf("%04d-01-01", year), fmt.Sprintf("%04d-12-31", year))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.MonthStat, 12)
	for i := range result {
		result[i].Month = i + 1
	}
	for rows.Next() {
		var s models.MonthStat
		if err := rows.Scan(&s.Month, &s.Total, &s.Count); err != nil {
			return nil, err
		}
		if s.Month >= 1 && s.Month <= 12 {
			result[s.Month-1] = s
		}
	}
	return result, rows.Err()
}

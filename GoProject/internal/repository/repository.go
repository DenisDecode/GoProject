package repository

import (
	"GoProject/internal/models"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Получение списка трат с учетом динамических фильтров и JOIN
func (r *Repository) GetFilteredExpenses(categoryID string, dateFrom, dateTo string) ([]models.Expense, error) {
	// Базовый запрос с JOIN таблицы expenses и categories
	query := `
        SELECT e.id, e.amount, e.description, e.date, e.category_id, c.name 
        FROM expenses e
        LEFT JOIN categories c ON e.category_id = c.id
        WHERE 1=1
    `
	var args []interface{}

	// Динамически добавляем условия WHERE, если фильтры переданы
	if categoryID != "" && categoryID != "0" {
		query += " AND e.category_id = ?"
		args = append(args, categoryID)
	}
	if dateFrom != "" {
		query += " AND e.date >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		query += " AND e.date <= ?"
		args = append(args, dateTo)
	}

	query += " ORDER BY e.date DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var exp models.Expense
		err := rows.Scan(&exp.ID, &exp.Amount, &exp.Description, &exp.Date, &exp.CategoryID, &exp.CategoryName)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}

	return expenses, nil
}

// Добавление новой траты (защита от SQL-инъекций через подготовленные запросы с ?)
func (r *Repository) CreateExpense(amount float64, description, date string, categoryID int) error {
	query := `INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, amount, description, date, categoryID)
	return err
}

// Получение всех категорий для выпадающего списка (<select>)
func (r *Repository) GetAllCategories() ([]models.Category, error) {
	rows, err := r.db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

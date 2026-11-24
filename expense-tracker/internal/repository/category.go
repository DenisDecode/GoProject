package repository

import (
	"context"
	"database/sql"

	"expense-tracker/internal/models"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *CategoryRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)`, id).Scan(&exists)
	return exists == 1, err
}

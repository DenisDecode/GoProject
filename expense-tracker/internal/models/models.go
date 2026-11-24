// Package models содержит структуры данных предметной области.
package models

import "errors"

var (
	ErrNotFound       = errors.New("запись не найдена")
	ErrDuplicateEmail = errors.New("пользователь с таким email уже существует")
)

// Expense — одна трата.
type Expense struct {
	ID           int64
	UserID       int64
	Amount       float64
	Description  string
	Date         string // формат ГГГГ-ММ-ДД
	CategoryID   int64  // 0 — без категории
	CategoryName string // заполняется через JOIN
}

type Category struct {
	ID   int64
	Name string
}

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    string
}

// ExpenseFilter — необязательные фильтры списка трат (пустое значение = фильтр не задан).
type ExpenseFilter struct {
	CategoryID int64
	From       string
	To         string
}

type CategoryStat struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

type MonthStat struct {
	Month int     `json:"month"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

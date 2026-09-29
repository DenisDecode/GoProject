package models

import "time"

// Expense описывает одну расходную операцию
type Expense struct {
	Amount      float64
	Description string
	Date        time.Time
}

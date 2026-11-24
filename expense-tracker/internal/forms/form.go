// Package forms — разбор и серверная валидация данных форм.
package forms

import "time"

const DateLayout = "2006-01-02"

// Base хранит ошибки валидации: имя поля -> текст ошибки.
type Base struct {
	Errors map[string]string
}

func (b *Base) AddError(field, message string) {
	if b.Errors == nil {
		b.Errors = make(map[string]string)
	}
	if _, exists := b.Errors[field]; !exists {
		b.Errors[field] = message
	}
}

func (b *Base) Valid() bool {
	return len(b.Errors) == 0
}

func IsValidDate(s string) bool {
	_, err := time.Parse(DateLayout, s)
	return err == nil
}

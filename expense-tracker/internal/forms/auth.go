package forms

import (
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"
)

type RegisterForm struct {
	Base
	Email    string
	Password string // никогда не выводится обратно в шаблон
	Confirm  string
}

func RegisterFormFromRequest(r *http.Request) *RegisterForm {
	return &RegisterForm{
		Base:     Base{Errors: map[string]string{}},
		Email:    strings.ToLower(strings.TrimSpace(r.PostFormValue("email"))),
		Password: r.PostFormValue("password"),
		Confirm:  r.PostFormValue("password_confirm"),
	}
}

func (f *RegisterForm) Validate() {
	validateEmail(&f.Base, f.Email)
	switch {
	case f.Password == "":
		f.AddError("password", "Введите пароль")
	case utf8.RuneCountInString(f.Password) < 8:
		f.AddError("password", "Пароль должен быть не короче 8 символов")
	case len(f.Password) > 72: // ограничение bcrypt
		f.AddError("password", "Пароль слишком длинный (максимум 72 байта)")
	}
	if f.Confirm != f.Password {
		f.AddError("password_confirm", "Пароли не совпадают")
	}
}

type LoginForm struct {
	Base
	Email    string
	Password string
}

func LoginFormFromRequest(r *http.Request) *LoginForm {
	return &LoginForm{
		Base:     Base{Errors: map[string]string{}},
		Email:    strings.ToLower(strings.TrimSpace(r.PostFormValue("email"))),
		Password: r.PostFormValue("password"),
	}
}

func (f *LoginForm) Validate() {
	if f.Email == "" {
		f.AddError("email", "Введите email")
	}
	if f.Password == "" {
		f.AddError("password", "Введите пароль")
	}
}

func validateEmail(b *Base, email string) {
	switch {
	case email == "":
		b.AddError("email", "Введите email")
	case len(email) > 254:
		b.AddError("email", "Слишком длинный email")
	default:
		addr, err := mail.ParseAddress(email)
		if err != nil || addr.Address != email {
			b.AddError("email", "Некорректный email")
		}
	}
}

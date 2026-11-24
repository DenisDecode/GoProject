package handlers

import (
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"expense-tracker/internal/forms"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/models"
)

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if currentUser(r) != nil {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		return
	}
	data := h.newTemplateData(r)
	data.Form = &forms.RegisterForm{}
	h.render(w, r, http.StatusOK, "register.html", data)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest)
		return
	}
	form := forms.RegisterFormFromRequest(r)
	form.Validate()
	if !form.Valid() {
		h.renderAuthForm(w, r, "register.html", form)
		return
	}

	// В БД сохраняется только bcrypt-хэш (с солью внутри), не пароль.
	hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	userID, err := h.users.Create(r.Context(), form.Email, string(hash))
	if errors.Is(err, models.ErrDuplicateEmail) {
		form.AddError("email", "Пользователь с таким email уже зарегистрирован")
		h.renderAuthForm(w, r, "register.html", form)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}

	if err := h.startSession(w, r, userID); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if currentUser(r) != nil {
		http.Redirect(w, r, "/expenses", http.StatusSeeOther)
		return
	}
	data := h.newTemplateData(r)
	data.Form = &forms.LoginForm{}
	h.render(w, r, http.StatusOK, "login.html", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest)
		return
	}
	form := forms.LoginFormFromRequest(r)
	form.Validate()
	if !form.Valid() {
		h.renderAuthForm(w, r, "login.html", form)
		return
	}

	user, err := h.users.GetByEmail(r.Context(), form.Email)
	if errors.Is(err, models.ErrNotFound) {
		// Сравниваем с фиктивным хэшем, чтобы время ответа не выдавало,
		// существует ли такой email.
		_ = bcrypt.CompareHashAndPassword(h.dummyHash, []byte(form.Password))
		form.AddError("form", "Неверный email или пароль")
		h.renderAuthForm(w, r, "login.html", form)
		return
	} else if err != nil {
		h.serverError(w, r, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(form.Password)); err != nil {
		form.AddError("form", "Неверный email или пароль")
		h.renderAuthForm(w, r, "login.html", form)
		return
	}

	if err := h.startSession(w, r, user.ID); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// Logout удаляет сессию из БД (инвалидирует) и стирает cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil && cookie.Value != "" {
		if err := h.sessions.Delete(r.Context(), cookie.Value); err != nil {
			log.Printf("удаление сессии: %v", err)
		}
	}
	middleware.ClearSessionCookie(w, h.cfg.CookieSecure)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, expires, err := h.sessions.Create(r.Context(), userID, h.cfg.SessionTTL)
	if err != nil {
		return err
	}
	middleware.SetSessionCookie(w, token, expires, h.cfg.CookieSecure)
	return nil
}

func (h *Handler) renderAuthForm(w http.ResponseWriter, r *http.Request, page string, form any) {
	data := h.newTemplateData(r)
	data.Form = form
	h.render(w, r, http.StatusUnprocessableEntity, page, data)
}

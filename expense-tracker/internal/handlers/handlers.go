// Package handlers — HTTP-обработчики. SQL здесь нет: только вызовы репозиториев.
package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"expense-tracker/internal/config"
	"expense-tracker/internal/middleware"
	"expense-tracker/internal/models"
	"expense-tracker/internal/render"
	"expense-tracker/internal/repository"
)

const maxFormSize = 1 << 20 // 1 МБ — защита от огромных тел запросов

type Handler struct {
	cfg        config.Config
	renderer   *render.Renderer
	expenses   *repository.ExpenseRepository
	categories *repository.CategoryRepository
	users      *repository.UserRepository
	sessions   *repository.SessionRepository
	dummyHash  []byte // для одинакового времени ответа при неверном email
}

func New(
	cfg config.Config,
	renderer *render.Renderer,
	expenses *repository.ExpenseRepository,
	categories *repository.CategoryRepository,
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
) *Handler {
	dummy, err := bcrypt.GenerateFromPassword([]byte("timing-attack-protection"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
	}
	return &Handler{
		cfg: cfg, renderer: renderer,
		expenses: expenses, categories: categories,
		users: users, sessions: sessions,
		dummyHash: dummy,
	}
}

type filterView struct {
	Category string
	From     string
	To       string
	Active   bool
}

// templateData — данные, которые получает любой шаблон.
type templateData struct {
	User         *models.User
	CurrentPath  string
	CurrentYear  int
	Expenses     []models.Expense
	Expense      *models.Expense
	Categories   []models.Category
	Filter       filterView
	FilterErrors []string
	Total        float64
	Count        int
	Years        []int
	Form         any
	Status       int
	StatusText   string
}

func (h *Handler) newTemplateData(r *http.Request) templateData {
	user, _ := middleware.UserFromContext(r.Context())
	return templateData{
		User:        user,
		CurrentPath: r.URL.Path,
		CurrentYear: time.Now().Year(),
	}
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	if err := h.renderer.Render(w, status, page, data); err != nil {
		h.serverError(w, r, err)
	}
}

// serverError логирует подробности, а пользователю показывает только страницу 500.
func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR %s %s: %v", r.Method, r.URL.Path, err)
	h.renderError(w, r, http.StatusInternalServerError)
}

func (h *Handler) renderError(w http.ResponseWriter, r *http.Request, status int) {
	data := h.newTemplateData(r)
	data.Status = status
	data.StatusText = http.StatusText(status)

	page := "500.html"
	if status == http.StatusNotFound {
		page = "404.html"
	}
	if err := h.renderer.Render(w, status, page, data); err != nil {
		log.Printf("не удалось отрендерить страницу ошибки: %v", err)
		http.Error(w, http.StatusText(status), status)
	}
}

// NotFound — кастомная страница 404.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.renderError(w, r, http.StatusNotFound)
}

// PanicResponse вызывается middleware Recover.
func (h *Handler) PanicResponse(w http.ResponseWriter, r *http.Request) {
	h.renderError(w, r, http.StatusInternalServerError)
}

func currentUser(r *http.Request) *models.User {
	user, _ := middleware.UserFromContext(r.Context())
	return user
}

func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"expense-tracker/internal/forms"
	"expense-tracker/internal/models"
)

func (h *Handler) StatsPage(w http.ResponseWriter, r *http.Request) {
	data := h.newTemplateData(r)
	for y := data.CurrentYear; y >= data.CurrentYear-5; y-- {
		data.Years = append(data.Years, y)
	}
	h.render(w, r, http.StatusOK, "stats.html", data)
}

type categoryStatsResponse struct {
	From  string                `json:"from"`
	To    string                `json:"to"`
	Total float64               `json:"total"`
	Items []models.CategoryStat `json:"items"`
}

type monthStatsResponse struct {
	Year  int                `json:"year"`
	Total float64            `json:"total"`
	Items []models.MonthStat `json:"items"`
}

// GET /api/stats/by-category?from=ГГГГ-ММ-ДД&to=ГГГГ-ММ-ДД
func (h *Handler) APIStatsByCategory(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	q := r.URL.Query()
	from := strings.TrimSpace(q.Get("from"))
	to := strings.TrimSpace(q.Get("to"))

	if from != "" && !forms.IsValidDate(from) {
		writeJSONError(w, http.StatusBadRequest, "параметр from должен быть в формате ГГГГ-ММ-ДД")
		return
	}
	if to != "" && !forms.IsValidDate(to) {
		writeJSONError(w, http.StatusBadRequest, "параметр to должен быть в формате ГГГГ-ММ-ДД")
		return
	}
	if from != "" && to != "" && from > to {
		writeJSONError(w, http.StatusBadRequest, "from не может быть позже to")
		return
	}

	items, err := h.expenses.SumByCategory(r.Context(), user.ID, from, to)
	if err != nil {
		log.Printf("ERROR stats by category: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	total, _, err := h.expenses.Summary(r.Context(), user.ID, models.ExpenseFilter{From: from, To: to})
	if err != nil {
		log.Printf("ERROR stats summary: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	writeJSON(w, http.StatusOK, categoryStatsResponse{From: from, To: to, Total: total, Items: items})
}

// GET /api/stats/by-month?year=2026
func (h *Handler) APIStatsByMonth(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	year := time.Now().Year()
	if v := strings.TrimSpace(r.URL.Query().Get("year")); v != "" {
		y, err := strconv.Atoi(v)
		if err != nil || y < 2000 || y > 2100 {
			writeJSONError(w, http.StatusBadRequest, "параметр year должен быть годом от 2000 до 2100")
			return
		}
		year = y
	}

	items, err := h.expenses.SumByMonth(r.Context(), user.ID, year)
	if err != nil {
		log.Printf("ERROR stats by month: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	total, _, err := h.expenses.Summary(r.Context(), user.ID, models.ExpenseFilter{
		From: strconv.Itoa(year) + "-01-01",
		To:   strconv.Itoa(year) + "-12-31",
	})
	if err != nil {
		log.Printf("ERROR stats summary: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	writeJSON(w, http.StatusOK, monthStatsResponse{Year: year, Total: total, Items: items})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		log.Printf("json: %v", err)
		status = http.StatusInternalServerError
		body = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

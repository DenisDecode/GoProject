// Package render загружает HTML-шаблоны (layout + страница) и рендерит их.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Renderer struct {
	templates map[string]*template.Template
}

var funcs = template.FuncMap{
	"money":      FormatMoney,
	"formatDate": formatDate,
}

// New компилирует каждую страницу вместе с общим layout.html один раз при старте.
func New(dir string) (*Renderer, error) {
	layout := filepath.Join(dir, "layout.html")

	var pages []string
	for _, pattern := range []string{"pages/*.html", "errors/*.html"} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		pages = append(pages, matches...)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("шаблоны не найдены в %q (запускайте сервер из корня проекта)", dir)
	}

	r := &Renderer{templates: make(map[string]*template.Template)}
	for _, page := range pages {
		name := filepath.Base(page)
		t, err := template.New(name).Funcs(funcs).ParseFiles(layout, page)
		if err != nil {
			return nil, fmt.Errorf("шаблон %s: %w", name, err)
		}
		r.templates[name] = t
	}
	return r, nil
}

// Render сначала рендерит в буфер: если в шаблоне ошибка, клиент не получит
// половину страницы, и обработчик сможет отдать корректную страницу 500.
func (r *Renderer) Render(w http.ResponseWriter, status int, page string, data any) error {
	t, ok := r.templates[page]
	if !ok {
		return fmt.Errorf("шаблон %q не найден", page)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w) // ошибка записи = клиент отключился, сделать уже ничего нельзя
	return nil
}

// FormatMoney: 1234.5 -> "1 234,50 ₽"
func FormatMoney(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	intPart, frac := s[:len(s)-3], s[len(s)-2:]
	var b strings.Builder
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteRune('\u00a0')
		}
		b.WriteRune(ch)
	}
	res := b.String() + "," + frac + "\u00a0₽"
	if neg {
		res = "-" + res
	}
	return res
}

func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return t.Format("02.01.2006")
}

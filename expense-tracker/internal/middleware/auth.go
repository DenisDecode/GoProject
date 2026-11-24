package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"expense-tracker/internal/models"
)

const SessionCookieName = "session_id"

// Неэкспортируемый тип ключа исключает коллизии с ключами других пакетов.
type contextKey string

const userContextKey contextKey = "user"

type SessionLookup interface {
	GetUserID(ctx context.Context, token string) (int64, error)
}

type UserLookup interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
}

// LoadUser читает cookie, проверяет сессию (включая срок действия) и,
// если всё в порядке, кладёт пользователя в context запроса.
func LoadUser(sessions SessionLookup, users UserLookup, secureCookie bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := sessions.GetUserID(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, models.ErrNotFound) {
					ClearSessionCookie(w, secureCookie) // сессия истекла или удалена
				} else {
					log.Printf("проверка сессии: %v", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			user, err := users.GetByID(r.Context(), userID)
			if err != nil {
				if errors.Is(err, models.ErrNotFound) {
					ClearSessionCookie(w, secureCookie)
				} else {
					log.Printf("загрузка пользователя: %v", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth пропускает только авторизованных. Страницы -> redirect на /login,
// API -> 401 с JSON.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"требуется авторизация"}`))
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok && user != nil
}

func SetSessionCookie(w http.ResponseWriter, token string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true, // недоступна из JavaScript
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

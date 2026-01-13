// internal/interfaces/middleware/csrf.go
package middleware

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// Envolve um http.Handler (ex: *gin.Engine)
func WrapWithCSRF(handler http.Handler) http.Handler {
	csrfHandler := nosurf.New(handler)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // true em produção (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})
	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "CSRF token inválido", http.StatusForbidden)
	}))
	return csrfHandler
}

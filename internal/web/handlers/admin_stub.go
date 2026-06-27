package handlers

import (
	"net/http"

	"github.com/jeb-maker/revues/internal/web/middleware"
)

// AdminStub is a placeholder until issue #9 (admin UI).
func AdminStub(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if user != nil {
		if _, err := w.Write([]byte("admin:" + user.Login)); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

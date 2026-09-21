package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"summary-your-usage/internal/usecase"
)

func NewRouter(users *usecase.User, groups *usecase.Group) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	h := &userHandler{users: users}
	g := &groupHandler{groups: groups}
	r.Route("/ui-api", func(r chi.Router) {
		r.Get("/users", h.list)
		r.Post("/users", h.create)
		r.Get("/users/{id}", h.get)
		r.Put("/users/{id}", h.update)
		r.Delete("/users/{id}", h.delete)
		r.Get("/groups", g.list)
		r.Post("/groups", g.create)
		r.Get("/groups/{id}", g.get)
		r.Put("/groups/{id}", g.update)
		r.Delete("/groups/{id}", g.delete)
	})
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "接口不存在")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "请求方法不支持")
	})
	return r
}

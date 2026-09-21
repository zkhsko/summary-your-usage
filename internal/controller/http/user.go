package http

import (
	"errors"
	"log"
	"net/http"

	"summary-your-usage/internal/entity"
	"summary-your-usage/internal/usecase"
)

type userHandler struct {
	users *usecase.User
}

func (h *userHandler) list(w http.ResponseWriter, r *http.Request) {
	result, err := h.users.List(r.Context())
	if err != nil {
		writeUserError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *userHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "用户")
	if !ok {
		return
	}
	user, err := h.users.Get(r.Context(), id)
	if err != nil {
		writeUserError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *userHandler) create(w http.ResponseWriter, r *http.Request) {
	var input entity.UserInput
	if !readJSON(w, r, &input) {
		return
	}
	user, err := h.users.Create(r.Context(), input)
	if err != nil {
		writeUserError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *userHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "用户")
	if !ok {
		return
	}
	var input entity.UserInput
	if !readJSON(w, r, &input) {
		return
	}
	user, err := h.users.Update(r.Context(), id, input)
	if err != nil {
		writeUserError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *userHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "用户")
	if !ok {
		return
	}
	if err := h.users.Delete(r.Context(), id); err != nil {
		writeUserError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeUserError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, entity.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user_not_found", "用户不存在")
	case errors.Is(err, entity.ErrEmailExists):
		writeError(w, http.StatusConflict, "email_exists", "该邮箱已被使用")
	case errors.Is(err, entity.ErrInvalidUser):
		writeError(w, http.StatusBadRequest, "validation_error", "请填写姓名（最多 100 个字符）和有效邮箱")
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用，请稍后重试")
	}
}

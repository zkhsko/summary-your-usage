package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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
	id, ok := userId(w, r)
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
	input, ok := readUserInput(w, r)
	if !ok {
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
	id, ok := userId(w, r)
	if !ok {
		return
	}
	input, ok := readUserInput(w, r)
	if !ok {
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
	id, ok := userId(w, r)
	if !ok {
		return
	}
	if err := h.users.Delete(r.Context(), id); err != nil {
		writeUserError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readUserInput(w http.ResponseWriter, r *http.Request) (entity.UserInput, bool) {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "请使用 application/json")
		return entity.UserInput{}, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input entity.UserInput
	err = decoder.Decode(&input)
	if err == nil {
		if err = decoder.Decode(new(any)); err == io.EOF {
			err = nil
		} else if err == nil {
			err = errors.New("multiple JSON values")
		}
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "请求内容过大")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_json", "请求 JSON 格式错误或包含未知字段")
		}
		return entity.UserInput{}, false
	}
	return input, true
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

func userId(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "validation_error", "用户 ID 必须为正整数")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

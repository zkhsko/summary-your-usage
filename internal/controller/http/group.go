package http

import (
	"errors"
	"log"
	"net/http"

	"summary-your-usage/internal/entity"
	"summary-your-usage/internal/usecase"
)

type groupHandler struct {
	groups *usecase.Group
}

func (h *groupHandler) list(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groups.List(r.Context())
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *groupHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "分组")
	if !ok {
		return
	}
	group, err := h.groups.Get(r.Context(), id)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, group)
}

func (h *groupHandler) create(w http.ResponseWriter, r *http.Request) {
	var input entity.GroupInput
	if !readJSON(w, r, &input) {
		return
	}
	group, err := h.groups.Create(r.Context(), input)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

func (h *groupHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "分组")
	if !ok {
		return
	}
	var input entity.GroupInput
	if !readJSON(w, r, &input) {
		return
	}
	group, err := h.groups.Update(r.Context(), id, input)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, group)
}

func (h *groupHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := resourceId(w, r, "分组")
	if !ok {
		return
	}
	if err := h.groups.Delete(r.Context(), id); err != nil {
		writeGroupError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeGroupError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, entity.ErrGroupNotFound):
		writeError(w, http.StatusNotFound, "group_not_found", "分组不存在")
	case errors.Is(err, entity.ErrGroupNameExists):
		writeError(w, http.StatusConflict, "group_name_exists", "该分组名已存在")
	case errors.Is(err, entity.ErrInvalidGroup):
		writeError(w, http.StatusBadRequest, "validation_error", "请填写分组名（最多 100 个字符）和有效的非负计费倍率，说明信息最多 1000 个字符")
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		writeError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用，请稍后重试")
	}
}

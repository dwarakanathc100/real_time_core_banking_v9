package customer

import (
	"encoding/json"
	"net/http"
)

type Handler struct{ repo *Repo }

func NewHandler(r *Repo) *Handler { return &Handler{repo: r} }

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	userID := 0
	if v := r.Context().Value("claims"); v != nil {
		// try to extract sub claim if possible
		// keep simple: set userID = 1
		userID = 1
	}
	var c Customer
	_ = json.NewDecoder(r.Body).Decode(&c)
	if c.FirstName == "" || c.Email == "" {
		http.Error(w, "email and password bad", http.StatusBadRequest)
		return
	}
	if err := h.repo.Create(&c, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

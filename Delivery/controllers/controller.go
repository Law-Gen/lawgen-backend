package controllers

import (
	"encoding/json"
	"lawgen/admin-service/Domain"
	usecases "lawgen/admin-service/Usecases"
	"net/http"
	"strconv"
	"strings"
)

type UserController struct {
	usecase *usecases.UserUsecase
}

func NewUserController (uc *usecases.UserUsecase) *UserController {
	return &UserController{usecase: uc}
}

func (c *UserController) UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.HandleGetAllUsers(w, r)
	case http.MethodPut:
		c.HandleUpdateUser(w, r)
	case http.MethodDelete:
		c.HandleDeleteUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *UserController) HandleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("serach")

	if page < 1 { page = 1}
	if limit < 1 { limit = 10}

	response, err := c.usecase.FetchAllUsers(r.Context(), page, limit, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *UserController) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	userID := parts[3]

	var payload Domain.UserUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.usecase.UpdateUser(r.Context(), userID, payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message":"User updated."})
}

func (c *UserController) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID := parts[3]
	if err := c.usecase.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
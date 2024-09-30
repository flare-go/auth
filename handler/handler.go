package handler

import (
	"encoding/json"
	"net/http"

	"goflare.io/auth/authentication"
	"goflare.io/auth/firebase"
	"goflare.io/auth/models/enum"
)

type UserHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	CheckPermission(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	authentication  authentication.Service
	firebaseService firebase.Service
}

func NewUserHandler(
	authentication authentication.Service,
	firebaseService firebase.Service,
) UserHandler {
	return &userHandler{
		authentication:  authentication,
		firebaseService: firebaseService,
	}
}

func (uh *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paseto, err := uh.authentication.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paseto)
}

func (uh *userHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		http.Error(w, "email or password is empty", http.StatusBadRequest)
		return
	}

	paseto, err := uh.authentication.Register(r.Context(), req.Username, req.Password, req.Email, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paseto)
}

func (uh *userHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uint32)

	ok, err := uh.authentication.CheckPermission(r.Context(), userID, enum.ResourceType(req.Resource), enum.ActionType(req.Action))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ok)
}

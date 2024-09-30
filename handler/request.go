package handler

// LoginRequest is the request for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest is the request for the register endpoint.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3"`
	Phone    string `json:"phone"`
}

// CheckPermissionRequest is the request for the check permission endpoint.
type CheckPermissionRequest struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

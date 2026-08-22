package structsv1

import "github.com/JayPonda/Product-catalog/server/src/models"

// RegisterRequest is the payload for creating a new account.
type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" validate:"required,min=1,max=50"`
	Email     string `json:"email" validate:"required,email,max=254"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest is the payload for authenticating.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=1,max=72"`
}

// UserResponse is the public user representation (no password).
type UserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// AuthResponse is returned after a successful register/login.
type AuthResponse struct {
	User UserResponse `json:"user"`
}

// ToUserResponse converts a model User into the public response DTO.
func ToUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}

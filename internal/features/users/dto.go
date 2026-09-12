package users

// CreateRequest is the payload for creating a user.
type CreateRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateRequest is the payload for updating a user.
type UpdateRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"required,email"`
}

// Response is the public shape of a user.
type Response struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

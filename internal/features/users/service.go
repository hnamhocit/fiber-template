package users

import (
	"context"
	"database/sql"
	"errors"
)

// ErrNotFound is returned when the requested user does not exist.
var ErrNotFound = errors.New("user not found")

// Service contains users business logic.
type Service struct {
	repo *Repository
}

// NewService creates a Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all users.
func (s *Service) List(ctx context.Context) ([]Response, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Response, len(rows))
	for i, r := range rows {
		out[i] = Response{ID: r.ID}
	}
	return out, nil
}

// GetByID returns a user by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*Response, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &Response{ID: row.ID}, nil
}

// Create creates a new user.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	row, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &Response{ID: row.ID}, nil
}

// Update updates a user.
func (s *Service) Update(ctx context.Context, id int64, req UpdateRequest) (*Response, error) {
	row, err := s.repo.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &Response{ID: row.ID}, nil
}

// Delete deletes a user.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

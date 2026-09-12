package users

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Row represents a row in the users table.
type Row struct {
	ID int64 `db:"id"`
}

// Repository provides access to the users table.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a Repository with the given DB.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// List returns all users.
func (r *Repository) List(ctx context.Context) ([]Row, error) {
	var rows []Row
	err := r.db.SelectContext(ctx, &rows,
		"SELECT id FROM users ORDER BY id")
	return rows, err
}

// GetByID returns a user by ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (*Row, error) {
	var row Row
	err := r.db.GetContext(ctx, &row,
		"SELECT id FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Create inserts a new user.
func (r *Repository) Create(ctx context.Context, req CreateRequest) (*Row, error) {
	var row Row
	err := r.db.QueryRowxContext(ctx,
		"INSERT INTO users DEFAULT VALUES RETURNING id").StructScan(&row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Update updates a user.
func (r *Repository) Update(ctx context.Context, id int64, req UpdateRequest) (*Row, error) {
	var row Row
	err := r.db.QueryRowxContext(ctx,
		"UPDATE users SET id = id WHERE id = $1 RETURNING id", id).StructScan(&row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Delete deletes a user.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM users WHERE id = $1", id)
	return err
}

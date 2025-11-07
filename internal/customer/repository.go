package customer

// Package customer handles customer database operations.

import (
	"database/sql"
	"time"
)

// Repo provides database methods for interacting with the customers table.
type Repo struct{ db *sql.DB }

// NewRepo initializes and returns a new customer repository instance.
func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

// Customer represents a customer entity in the system.
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Mobile    string    `json:"mobile"`
	CreatedAt time.Time `json:"created_at"`
}

// Create inserts a new customer record in the database.
func (r *Repo) Create(c *Customer, userID int) error {
	query := `
		INSERT INTO customers(user_id, caf, first_name, last_name, email, mobile)
		VALUES ($1, '{}', $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		userID,
		c.FirstName,
		c.LastName,
		c.Email,
		c.Mobile,
	).Scan(&c.ID, &c.CreatedAt)
}

// List retrieves the most recent 100 customers from the database.
func (r *Repo) List() ([]*Customer, error) {
	rows, err := r.db.Query("SELECT id, first_name,last_name,email,mobile,created_at FROM customers ORDER BY id DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Customer
	for rows.Next() {
		c := &Customer{}
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Mobile, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

package account

import (
	"database/sql"
	"time"
)

type Repo struct {
	db *sql.DB
}

func NewHandler(r *Repo) *Handler { return &Handler{repo: r} }
func NewRepo(db *sql.DB) *Repo    { return &Repo{db: db} }

type Account struct {
	ID            int       `json:"id"`
	CustomerID    int       `json:"customer_id"`
	AccountNumber string    `json:"account_number"`
	Currency      string    `json:"currency"`
	Balance       float64   `json:"balance"`
	CreatedAt     time.Time `json:"created_at"`
}

func (r *Repo) Create(a *Account) error {
	return r.db.QueryRow("INSERT INTO accounts(customer_id, account_number, currency, balance) VALUES($1,$2,$3,$4) RETURNING id, created_at", a.CustomerID, a.AccountNumber, a.Currency, a.Balance).Scan(&a.ID, &a.CreatedAt)
}

func (r *Repo) Get(id int) (*Account, error) {
	a := &Account{}
	if err := r.db.QueryRow("SELECT id, customer_id, account_number, currency, balance, created_at FROM accounts WHERE id=$1", id).
		Scan(&a.ID, &a.CustomerID, &a.AccountNumber, &a.Currency, &a.Balance, &a.CreatedAt); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) GetByAccountNumber(acct string) (*Account, error) {
	a := &Account{}
	if err := r.db.QueryRow("SELECT id, customer_id, account_number, currency, balance, created_at FROM accounts WHERE account_number=$1", acct).
		Scan(&a.ID, &a.CustomerID, &a.AccountNumber, &a.Currency, &a.Balance, &a.CreatedAt); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repo) UpdateBalanceTx(tx *sql.Tx, accountID int, newBal float64) error {
	_, err := tx.Exec("UPDATE accounts SET balance=$1 WHERE id=$2", newBal, accountID)
	return err
}

func (r *Repo) CreateTransactionTx(tx *sql.Tx, accountID int, related int, amount float64, typ, narr string) (int, error) {
	var id int
	err := tx.QueryRow("INSERT INTO transactions(account_id, related_account_id, amount, type, narration) VALUES($1,$2,$3,$4,$5) RETURNING id", accountID, related, amount, typ, narr).Scan(&id)
	return id, err
}

func (r *Repo) ListAccountsByCustomer(customerID int) ([]*Account, error) {
	rows, err := r.db.Query("SELECT id,customer_id,account_number,currency,balance,created_at FROM accounts WHERE customer_id=$1", customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Account
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.AccountNumber, &a.Currency, &a.Balance, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

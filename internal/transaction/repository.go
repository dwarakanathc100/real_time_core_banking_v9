package transaction

import (
    "database/sql"
    "time"
)

type Repo struct { db *sql.DB }
func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

type Transaction struct {
    ID int `json:"id"`
    AccountID int `json:"account_id"`
    Amount float64 `json:"amount"`
    Type string `json:"type"`
    Narration string `json:"narration"`
    CreatedAt time.Time `json:"created_at"`
}

func (r *Repo) ListForAccount(accountID int, from, to string) ([]*Transaction, error) {
    rows, err := r.db.Query("SELECT id,account_id,amount,type,narration,created_at FROM transactions WHERE account_id=$1 AND created_at BETWEEN $2 AND $3 ORDER BY created_at DESC", accountID, from, to)
    if err!=nil { return nil, err }
    defer rows.Close()
    var out []*Transaction
    for rows.Next() {
        t := &Transaction{}
        if err := rows.Scan(&t.ID,&t.AccountID,&t.Amount,&t.Type,&t.Narration,&t.CreatedAt); err!=nil { return nil, err }
        out = append(out, t)
    }
    return out, nil
}

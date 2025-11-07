// Package customer provides HTTP handlers for managing customers.
// Package account also includes HTTP handlers for account operations such as
// creating accounts, checking balance, deposits, withdrawals, and transfers.
package account

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Handler manages HTTP requests related to account operations.
type Handler struct{ repo *Repo }

// CreateCustomer handles POST /v1/customers to create a new customer.
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var a Account
	_ = json.NewDecoder(r.Body).Decode(&a)
	logrus.Infof("CreateAccount %v", a)
	if a.CustomerID == 0 || a.AccountNumber == "" {
		http.Error(w, "customer info. missing", http.StatusBadRequest)
		return
	}
	if err := h.repo.Create(&a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(a)
}

// GetBalance handles GET /v1/accounts/balance to fetch the balance for a given account number.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	acct := q.Get("account_number")
	if acct == "" {
		http.Error(w, "missing", http.StatusBadRequest)
		return
	}
	a, err := h.repo.GetByAccountNumber(acct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]float64{"balance": a.Balance})
}

// Deposit handles POST /v1/accounts/deposit to deposit an amount into an account.
func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	type req struct {
		AccountNumber string  `json:"account_number"`
		Amount        float64 `json:"amount"`
	}
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	if rr.AccountNumber == "" || rr.Amount <= 0 {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	// simple transactional update
	tx, err := h.repo.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a, err := h.repo.GetByAccountNumber(rr.AccountNumber)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	newBal := a.Balance + rr.Amount
	if err := h.repo.UpdateBalanceTx(tx, a.ID, newBal); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := h.repo.CreateTransactionTx(tx, a.ID, 0, rr.Amount, "deposit", "deposit"); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Infof("deposited %.2f to %s", rr.Amount, rr.AccountNumber)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Withdraw handles POST /v1/accounts/withdraw to withdraw funds from an account.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	type req struct {
		AccountNumber string  `json:"account_number"`
		Amount        float64 `json:"amount"`
	}
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	if rr.AccountNumber == "" || rr.Amount <= 0 {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	tx, err := h.repo.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a, err := h.repo.GetByAccountNumber(rr.AccountNumber)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if a.Balance < rr.Amount {
		tx.Rollback()
		http.Error(w, "insufficient", http.StatusBadRequest)
		return
	}
	newBal := a.Balance - rr.Amount
	if err := h.repo.UpdateBalanceTx(tx, a.ID, newBal); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := h.repo.CreateTransactionTx(tx, a.ID, 0, rr.Amount, "withdraw", "withdraw"); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Transfer handles POST /v1/accounts/transfer to move funds between two accounts.
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	type req struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	}
	var rr req
	_ = json.NewDecoder(r.Body).Decode(&rr)
	if rr.From == "" || rr.To == "" || rr.Amount <= 0 {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	// use db transaction to make atomic transfer
	tx, err := h.repo.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fromAcc, err := h.repo.GetByAccountNumber(rr.From)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	toAcc, err := h.repo.GetByAccountNumber(rr.To)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if fromAcc.Balance < rr.Amount {
		tx.Rollback()
		http.Error(w, "insufficient", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateBalanceTx(tx, fromAcc.ID, fromAcc.Balance-rr.Amount); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.repo.UpdateBalanceTx(tx, toAcc.ID, toAcc.Balance+rr.Amount); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := h.repo.CreateTransactionTx(tx, fromAcc.ID, toAcc.ID, rr.Amount, "transfer_debit", "transfer out"); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := h.repo.CreateTransactionTx(tx, toAcc.ID, fromAcc.ID, rr.Amount, "transfer_credit", "transfer in"); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

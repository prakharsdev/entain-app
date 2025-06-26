package user

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"entain-app/internal/db"
	"entain-app/pkg/utils"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidAmount        = errors.New("invalid amount format")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
)

func ProcessTransaction(userID uint64, req TransactionRequest, sourceType string) error {
	// Validate amount
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		return ErrInvalidAmount
	}

	// Check for duplicate transaction ID
	var exists bool
	err = db.DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM transactions WHERE transaction_id = $1)`, req.TransactionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check transaction ID: %w", err)
	}
	if exists {
		return ErrDuplicateTransaction
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin db tx: %w", err)
	}
	defer tx.Rollback()

	// Get current user balance
	var currentBalance float64
	err = tx.QueryRow(`SELECT balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("failed to fetch user balance: %w", err)
	}

	// Recalculate balance
	if req.State == "win" {
		currentBalance += amount
	} else if req.State == "lose" {
		if currentBalance < amount {
			return ErrInsufficientBalance
		}
		currentBalance -= amount
	} else {
		return errors.New("invalid state value")
	}

	// Update balance
	_, err = tx.Exec(`UPDATE users SET balance = $1 WHERE id = $2`, currentBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Insert transaction
	_, err = tx.Exec(`
		INSERT INTO transactions (transaction_id, user_id, amount, state, source_type)
		VALUES ($1, $2, $3, $4, $5)`,
		req.TransactionID, userID, amount, req.State, sourceType)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	utils.Logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"transaction_id": req.TransactionID,
		"amount":         amount,
		"state":          req.State,
		"source_type":    sourceType,
	}).Info("Processed transaction")

	return nil
}

func GetUserBalance(userID uint64) (*User, error) {
	var u User
	err := db.DB.QueryRow(`SELECT id, balance FROM users WHERE id = $1`, userID).Scan(&u.ID, &u.Balance)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}
	return &u, nil
}

package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"entain-app/pkg/utils"

	"github.com/gorilla/mux"
)

// HandleTransaction processes incoming transactions with idempotency.
func HandleTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil || userID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Validate Source-Type header
	sourceType := r.Header.Get("Source-Type")
	if sourceType == "" || !utils.IsValidSourceType(sourceType) {
		utils.WriteError(w, http.StatusBadRequest, "Missing or invalid Source-Type header")
		return
	}

	// Parse and validate request body
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	
	// Validate transaction state and amount precision
	if !utils.IsValidState(req.State) {
		utils.WriteError(w, http.StatusBadRequest, "Invalid state: must be 'win' or 'lose'")
		return
	}

	if !utils.IsValidAmountFormat(req.Amount) {
		utils.WriteError(w, http.StatusBadRequest, "Amount must have at most 2 decimal places")
		return
	}

	// Process the transaction
	err = ProcessTransaction(userID, req, sourceType)
	switch err {
	case nil:
		utils.WriteSuccess(w, http.StatusOK, TransactionResponse{Message: "Transaction processed"})
	case ErrDuplicateTransaction:
		utils.WriteSuccess(w, http.StatusOK, TransactionResponse{Message: "Transaction already processed"})
	case ErrInvalidAmount, ErrInsufficientBalance:
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	case ErrUserNotFound:
		utils.WriteError(w, http.StatusNotFound, err.Error())
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// HandleBalance returns the current balance for the given user.
func HandleBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil || userID == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := GetUserBalance(userID)
	if err != nil {
		if err == ErrUserNotFound {
			utils.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve balance")
		}
		return
	}

	resp := BalanceResponse{
		UserID:  user.ID,
		Balance: strconv.FormatFloat(user.Balance, 'f', 2, 64),
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

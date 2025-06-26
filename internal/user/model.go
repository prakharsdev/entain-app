package user

type User struct {
	ID      uint64  `json:"userId"`
	Balance float64 `json:"balance"`
}

type Transaction struct {
	TransactionID string  `json:"transactionId"`
	UserID        uint64  `json:"userId"`
	Amount        float64 `json:"amount"`
	State         string  `json:"state"`
	SourceType    string  `json:"sourceType"`
}

type TransactionRequest struct {
	State         string `json:"state"`         // win or lose
	Amount        string `json:"amount"`        // as string (e.g., "10.15")
	TransactionID string `json:"transactionId"` // must be unique
}

type TransactionResponse struct {
	Message string `json:"message"`
}

type BalanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"` // as string with 2 decimals
}

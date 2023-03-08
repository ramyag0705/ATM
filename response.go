package responses

type CreateAccount struct {
	Name string `json:"name" binding:"required"`
	Pin  string `json:"pin" binding:"required"`
}

type Deposit struct {
	AccountNumber string  `json:"account_number" binding:"required"`
	Pin           string  `json:"pin" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

type Withdraw struct {
	AccountNumber string  `json:"account_number" binding:"required"`
	Pin           string  `json:"pin" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
}

type Transfer struct {
	FromAccount string  `json:"from_account" binding:"required"`
	FromPin     string  `json:"from_pin" binding:"required"`
	ToAccount   string  `json:"to_account" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
}

type Pin struct {
	AccountNumber string `json:"account_number" binding:"required"`
	OldPin        string `json:"old_pin" binding:"required"`
	NewPin        string `json:"new_pin" binding:"required"`
}


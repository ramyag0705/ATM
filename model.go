package models

type Account struct {
	AccountNumber string  `bson:"account_number" json:"account_number"`
	Name          string  `bson:"name" json:"name"`
	Pin           string  `bson:"pin" json:"-"`
	Balance       float64 `bson:"balance" json:"balance"`
}

type Transaction struct {
	From     string  `bson:"from" json:"from"`
	To       string  `bson:"to" json:"to"`
	Type     string  `bson:"type" json:"type"`
	Amount   float64 `bson:"amount" json:"amount"`
	DateTime string  `bson:"datetime" json:"datetime"`
}

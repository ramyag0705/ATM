package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"account/configs"
	"account/models"
	"account/responses"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var client = configs.DB

func hashPassword(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}



func CreateAccount(c *gin.Context) {
	var req responses.CreateAccount
	if err := c.ShouldBindJSON(&req); err != nil {						// create an user account
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accountNumber := fmt.Sprintf("%06d", rand.Intn(1000000))

	
	filter := bson.M{"account_number": accountNumber}					// Check if the account exists
	var existingAccount models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filter).Decode(&existingAccount)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account already exists"})
		return
	}

	if len(req.Pin) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pin must be 4 digits"})
		return
	}

	
	newAccount := models.Account{										// Insert new account
		Name:          req.Name,
		AccountNumber: accountNumber,
		Pin:           hashPassword(req.Pin),
		Balance:       0,
	}
	_, err = client.Database("mydb").Collection("users").InsertOne(context.Background(), newAccount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Account created successfully. Account number": accountNumber})
}



func Deposit(c *gin.Context) {											// Depositing money
	var req responses.Deposit
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	filter := bson.M{"account_number": req.AccountNumber, "pin": hashPassword(req.Pin)}			// Check if account exists and PIN matches
	var existingAccount models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filter).Decode(&existingAccount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number or PIN"})
		return
	}

	
	existingAccount.Balance += req.Amount									// Update account balance
	update := bson.M{"$set": bson.M{"balance": existingAccount.Balance}}
	_, err = client.Database("mydb").Collection("users").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deposit money"})
		return
	}

	
	transaction := models.Transaction{										// Insert transaction record
		From:     "",
		To:       req.AccountNumber,
		Type:     "deposit",
		Amount:   req.Amount,
		DateTime: time.Now().Format(time.RFC3339),
	}
	_, err = client.Database("mydb").Collection("transactions").InsertOne(context.Background(), transaction)
	if err != nil {
		log.Println("failed to insert transaction record:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "money deposited successfully"})
}



func Withdraw(c *gin.Context) {												// withdraw money
	var req responses.Withdraw
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	filter := bson.M{"account_number": req.AccountNumber, "pin": hashPassword(req.Pin)}		//check if account exists and pin match
	var existingAccount models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filter).Decode(&existingAccount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number or PIN"})
		return
	}

	
	if existingAccount.Balance < req.Amount {												// Check if 'from' account has enough balance
		c.JSON(http.StatusBadRequest, gin.H{"error": "not enough balance in account"})
		return
	}

	
	existingAccount.Balance -= req.Amount													//update account balance
	update := bson.M{"$set": bson.M{"balance": existingAccount.Balance}}
	_, err = client.Database("mydb").Collection("users").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to withdraw money"})
		return
	}

	
	transaction := models.Transaction{														//insert transaction record
		From:     "",
		To:       req.AccountNumber,
		Type:     "withdraw",
		Amount:   req.Amount,
		DateTime: time.Now().Format(time.RFC3339),
	}
	_, err = client.Database("mydb").Collection("transactions").InsertOne(context.Background(), transaction)
	if err != nil {
		log.Println("failed to insert transaction record:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "money withdraw successful"})

}



func Transfer(c *gin.Context) {															// transfer money from one account to another
	var req responses.Transfer
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	filterFrom := bson.M{"account_number": req.FromAccount, "pin": hashPassword(req.FromPin)}		// Check if 'from' account exists and PIN matches
	var fromAccount models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filterFrom).Decode(&fromAccount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' account number or PIN"})
		return
	}

	
	filterTo := bson.M{"account_number": req.ToAccount}								// Check if 'to' account exists
	var toAccount models.Account
	err = client.Database("mydb").Collection("users").FindOne(context.Background(), filterTo).Decode(&toAccount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' account number"})
		return
	}

	
	if fromAccount.Balance < req.Amount {											// Check if 'from' account has enough balance
		c.JSON(http.StatusBadRequest, gin.H{"error": "not enough balance in 'from' account"})
		return
	}

	
	fromAccount.Balance -= req.Amount												// Update 'from' account balance
	updateFrom := bson.M{"$set": bson.M{"balance": fromAccount.Balance}}
	_, err = client.Database("mydb").Collection("users").UpdateOne(context.Background(), filterFrom, updateFrom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transfer money"})
		return
	}

	
	toAccount.Balance += req.Amount													// Update 'to' account balance
	updateTo := bson.M{"$set": bson.M{"balance": toAccount.Balance}}
	_, err = client.Database("mydb").Collection("users").UpdateOne(context.Background(), filterTo, updateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transfer money"})
		return
	}

	
	transactionFrom := models.Transaction{											// Insert transaction records
		From:     req.FromAccount,
		To:       req.ToAccount,
		Type:     "withdraw",
		Amount:   req.Amount,
		DateTime: time.Now().Format(time.RFC3339),
	}
	_, err = client.Database("mydb").Collection("transactions").InsertOne(context.Background(), transactionFrom)
	if err != nil {
		log.Println("failed to insert transaction record:", err)
	}

	transactionTo := models.Transaction{
		From:     req.FromAccount,
		To:       req.ToAccount,
		Type:     "deposit",
		Amount:   req.Amount,
		DateTime: time.Now().Format(time.RFC3339),
	}
	_, err = client.Database("mydb").Collection("transactions").InsertOne(context.Background(), transactionTo)
	if err != nil {
		log.Println("failed to insert transaction record:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "money transferred successfully"})
}

func SetPin(c *gin.Context) {
	var req responses.Pin
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	filter := bson.M{"account_number": req.AccountNumber, "pin": hashPassword(req.OldPin)}			// Check if account exists and old PIN matches
	var account models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number or PIN"})
		return
	}

	
	update := bson.M{"$set": bson.M{"pin": hashPassword(req.NewPin)}}								// Update PIN
	_, err = client.Database("mydb").Collection("users").UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update PIN"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PIN updated successfully"})
}

func BankStatement(c *gin.Context) {
	var req struct {
		AccountNumber string `json:"account_number"`
		Pin           string `json:"pin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	
	filter := bson.M{"account_number": req.AccountNumber}					// Find account by account number
	var account models.Account
	err := client.Database("mydb").Collection("users").FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account number"})
		return
	}

	
	if account.Pin != hashPassword(req.Pin) {								// Validate PIN
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid PIN"})
		return
	}

	
	filter = bson.M{"$or": []interface{}{									// Find all transactions for account
		bson.M{"from": req.AccountNumber},
		bson.M{"to": req.AccountNumber},
	}}
	cursor, err := client.Database("mydb").Collection("transactions").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve transaction history"})
		return
	}
	defer cursor.Close(context.Background())

	
	var transactions []models.Transaction								// Extract transactions from cursor
	for cursor.Next(context.Background()) {
		var transaction models.Transaction
		if err := cursor.Decode(&transaction); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode transaction"})
			return
		}
		transactions = append(transactions, transaction)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

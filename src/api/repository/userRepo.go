package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	model "github.com/myrachanto/entaingo/src/api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Userrepository repository
var (
	Userrepository UserrepoInterface = &userrepository{}
	Userrepo                         = userrepository{}
)

type Key struct {
	EncryptionKey string `mapstructure:"EncryptionKey"`
}

type UserrepoInterface interface {
	Create(transaction *model.TransactionRequest) (*model.UserInfo, error)
	CancelOddTransactions(ctx context.Context, wg *sync.WaitGroup)
	GetTransactions(userId int) (*model.UserInfo, error)
}
type userrepository struct{}

func NewUserRepo() UserrepoInterface {
	return &userrepository{}
}
func (r *userrepository) Create(transactionReq *model.TransactionRequest) (*model.UserInfo, error) {
	// Get default user
	defaultUser, err := r.GetUser()
	if err != nil {
		return nil, err
	}

	// Connect to the database
	gormdb, err := IndexRepo.Getconnected()
	if err != nil {
		return nil, err
	}
	defer IndexRepo.DbClose(gormdb)

	// Start transaction
	tx := gormdb.Begin()

	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if transaction already exists
	var existingTransaction model.Transaction
	if err := tx.Where("transaction_id = ?", transactionReq.TransactionID).First(&existingTransaction).Error; err == nil {
		// Transaction already processed
		tx.Rollback()
		return nil, fmt.Errorf("transaction already processed")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		return nil, handleError(tx, err, "failed to check existing transaction")
	}

	// Lock the user row for update (optimistic locking)
	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, defaultUser.ID).Error; err != nil {
		return nil, handleError(tx, err, "user not found")
	}

	// Update balance
	newBalance := user.Balance
	if transactionReq.State == "win" {
		newBalance += transactionReq.Amount
	} else if transactionReq.State == "lost" {
		newBalance -= transactionReq.Amount
		if newBalance < 0 {
			tx.Rollback()
			return nil, fmt.Errorf("balance cannot be negative")
		}
	}

	// Optimistic lock based on version
	if err := tx.Model(&user).Where("version = ?", user.Version).Updates(map[string]interface{}{
		"balance": newBalance,
		"version": user.Version + 1,
	}).Error; err != nil {
		return nil, handleError(tx, err, "failed to update balance, version conflict")
	}

	// Create transaction record
	transaction := model.Transaction{
		TransactionID: transactionReq.TransactionID,
		Amount:        transactionReq.Amount,
		State:         transactionReq.State,
		SourceType:    transactionReq.SourceType,
		UserID:        user.ID,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		return nil, handleError(tx, err, "failed to save transaction")
	}

	// Commit the transaction
	tx.Commit()

	// Return user info and transaction details
	return &model.UserInfo{
		User:        user,
		Transaction: []model.Transaction{transaction},
	}, nil
}

func handleError(tx *gorm.DB, err error, msg string) error {
	if err != nil {
		tx.Rollback()
		return fmt.Errorf(msg, err)
	}
	return nil
}

func (r *userrepository) CancelOddTransactions(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("CancelOddTransactions started ....")

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in routes ", err)
	}

	OddCancelInterval := os.Getenv("OddCancelInterval")
	OddInterval, err := strconv.ParseUint(OddCancelInterval, 10, 32)
	if err != nil {
		log.Fatal("failed to parse the odd Interval ", err)
	}

	gormdb, err := IndexRepo.Getconnected()
	if err != nil {
		log.Fatal("failed to connect to the database ", err)
	}
	defer IndexRepo.DbClose(gormdb)

	ticker := time.NewTicker(time.Minute * time.Duration(OddInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("N odd time cancellation initialized")

			// Select 10 latest odd transactions that haven't been canceled
			var transactions []model.Transaction
			if err := gormdb.Where("id % 2 != 0 AND canceled = ?", false).
				Order("processed_at desc").
				Limit(10).
				Find(&transactions).Error; err != nil {
				log.Println("Error fetching transactions: ", err)
				continue
			}

			// Use a transaction to cancel and update user balances in one batch
			tx := gormdb.Begin()

			for _, transaction := range transactions {
				var user model.User
				if err := tx.First(&user, transaction.UserID).Error; err != nil {
					tx.Rollback()
					log.Println("Failed to fetch user for transaction: ", err)
					continue
				}

				// Reverse balance impact
				if transaction.State == "win" {
					user.Balance -= transaction.Amount
				} else if transaction.State == "lost" {
					user.Balance += transaction.Amount
				}

				// Prevent negative balances
				if user.Balance < 0 {
					tx.Rollback()
					log.Println("Balance would be negative, skipping")
					continue
				}

				// Update user balance
				if err := tx.Save(&user).Error; err != nil {
					tx.Rollback()
					log.Println("Failed to update user balance: ", err)
					continue
				}

				// Mark transaction as canceled
				transaction.Canceled = true
				if err := tx.Save(&transaction).Error; err != nil {
					tx.Rollback()
					log.Println("Failed to cancel transaction: ", err)
					continue
				}
			}

			// Commit the transaction
			tx.Commit()

		case <-ctx.Done():
			log.Println("CancelOddTransactions gracefully shutting down...")
			return
		}
	}
}

func (r userrepository) GetTransactions(userId int) (*model.UserInfo, error) {
	gorm, err := IndexRepo.Getconnected()
	if err != nil {
		return nil, err
	}
	defer IndexRepo.DbClose(gorm)
	results := []model.Transaction{}
	errs := gorm.Find(&results).Error
	if errs != nil {
		return nil, fmt.Errorf("no results found %w", errs)
	}
	var user model.User
	errs = gorm.Where("id = ?", userId).First(&user).Error
	if errs != nil {
		return nil, fmt.Errorf("no results found %w", errs)
	}
	return &model.UserInfo{
		User:        user,
		Transaction: results,
	}, nil
}
func (r userrepository) GetUser() (*model.User, error) {
	gorm, err := IndexRepo.Getconnected()
	if err != nil {
		return nil, err
	}
	defer IndexRepo.DbClose(gorm)
	result := &model.User{}
	errs := gorm.First(&result).Error
	if errs != nil {
		return nil, fmt.Errorf("user not found %w", errs)
	}
	return result, err
}

func (r userrepository) transactionIdExist(transactionId string) bool {
	gorm, err := IndexRepo.Getconnected()
	if err != nil {
		return false
	}
	defer IndexRepo.DbClose(gorm)
	result := &model.Transaction{}
	errs := gorm.Where("transaction_id = ?", transactionId).First(&result).Error
	return errs == nil
}

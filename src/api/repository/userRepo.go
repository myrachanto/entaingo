package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	model "github.com/myrachanto/entaingo/src/api/models"
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

	gorm, err := IndexRepo.Getconnected()
	if err != nil {
		return nil, err
	}
	defer IndexRepo.DbClose(gorm)
	// check if the transactionid is unique
	if ok := r.transactionIdExist(transactionReq.TransactionID); ok {
		return nil, fmt.Errorf("transaction already processed")
	}

	// DB transaction
	tx := gorm.Begin()

	var user model.User
	// check if user exists
	if err := tx.First(&user, 1).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user not found")
	}

	var transaction model.Transaction

	// Calculate the new balance
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
	// Update user balance
	user.Balance = newBalance
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update balance")
	}

	// Save transaction
	transaction = model.Transaction{
		TransactionID: transactionReq.TransactionID,
		Amount:        transactionReq.Amount,
		State:         transactionReq.State,
		SourceType:    transactionReq.SourceType,
		UserID:        user.ID,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save transaction")
	}

	tx.Commit()
	return &model.UserInfo{
		User:        user,
		Transaction: []model.Transaction{transaction},
	}, nil

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

	gorm, err := IndexRepo.Getconnected()
	if err != nil {
		log.Fatal("failed to connect to the database ", err)
	}
	defer IndexRepo.DbClose(gorm)

	ticker := time.NewTicker(time.Minute * time.Duration(OddInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("N odd time cancellation initialized")
			// Select 10 latest odd transactions that haven't been canceled
			var transactions []model.Transaction
			gorm.Where("id % 2 != 0 AND canceled = ?", false).
				Order("processed_at desc").
				Limit(10).
				Find(&transactions)

			for _, transaction := range transactions {
				tx := gorm.Begin()
				var user model.User
				if err := tx.First(&user, transaction.UserID).Error; err != nil {
					tx.Rollback()
					continue
				}

				// Reverse balance impact
				if transaction.State == "win" {
					user.Balance -= transaction.Amount
				} else if transaction.State == "lost" {
					user.Balance += transaction.Amount
				}

				if user.Balance < 0 {
					tx.Rollback()
					continue
				}

				// Update user balance
				if err := tx.Save(&user).Error; err != nil {
					tx.Rollback()
					continue
				}

				// Mark transaction as canceled
				transaction.Canceled = true
				if err := tx.Save(&transaction).Error; err != nil {
					tx.Rollback()
					continue
				}

				tx.Commit()
			}
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

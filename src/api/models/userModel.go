package models

import "time"

type User struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	Balance float64 `gorm:"not null;default:0.00" json:"balance"` // Removed explicit type
	Version int     `gorm:"type:int;default:1"`                   // Optimistic locking
}

type Transaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TransactionID string    `gorm:"unique;not null" json:"transaction_id"`
	Amount        float64   `gorm:"not null" json:"amount"` // Removed explicit type
	State         string    `gorm:"type:varchar(10);not null" json:"state"`
	SourceType    string    `gorm:"type:varchar(50);not null" json:"source_type"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	ProcessedAt   time.Time `gorm:"autoCreateTime" json:"processed_at"` // Automatically set to current time
	Canceled      bool      `gorm:"default:false" json:"canceled"`
}

// type User struct {
// 	ID      uint    `gorm:"primaryKey" json:"id"`
// 	Balance float64 `gorm:"type:decimal(10,2);not null;default:0.00" json:"balance"`

// }

//	type Transaction struct {
//		ID            uint      `gorm:"primaryKey" json:"id"`
//		TransactionID string    `gorm:"unique;not null" json:"transaction_id"`
//		Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
//		State         string    `gorm:"type:varchar(10);not null" json:"state"`
//		SourceType    string    `gorm:"type:varchar(50);not null" json:"source_type"`
//		UserID        uint      `gorm:"not null" json:"user_id"`
//		ProcessedAt   time.Time `gorm:"autoCreateTime" json:"processed_at"`
//		Canceled      bool      `gorm:"default:false" json:"canceled"`
//	}
type TransactionRequest struct {
	State         string  `json:"state" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	TransactionID string  `json:"transactionId" binding:"required"`
	SourceType    string  `json:"source_type"`
}
type UserInfo struct {
	User        User          `json:"user"`
	Transaction []Transaction `json:"transaction"`
}

// Balance float64 `gorm:"type:decimal(10,2);not null;default:0.00" json:"balance"`

// Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`

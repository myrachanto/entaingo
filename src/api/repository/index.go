package repository

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	model "github.com/myrachanto/entaingo/src/api/models"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

// IndexRepo
var (
	IndexRepo indexRepo = indexRepo{}
)

// Layout ...
const (
	Layout   = "2006-01-02"
	layoutUS = "January 2, 2006"
)

type DbConfig struct {
	Host     string
	User     string
	Password string
	Dbname   string
	Port     string
	Timezone string
}

func LoaddbConfig() (DbConfig, error) {
	var db DbConfig

	// Set config file name (without extension) and type
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".") // Look for config in the working directory

	viper.AutomaticEnv() // Use environment variables if available

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return db, fmt.Errorf("failed to read configuration: %w", err)
	}

	// Unmarshal the config into the struct
	if err := viper.Unmarshal(&db); err != nil {
		return db, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return db, nil
}

// /curtesy to gorm
type indexRepo struct {
	Bizname string `json:"bizname,omitempty"`
}

func (indexRepo indexRepo) Dbsetup() error {
	// Load the DB configuration using Viper
	fmt.Println("step1 ...................................................")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in routes ", err)
	}

	fmt.Println("step2 ...................................................")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_TIMEZONE"),
	)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,  // Connection string for Postgres
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		SkipDefaultTransaction: true,                                // Skip default transactions for performance
		PrepareStmt:            true,                                // Caches prepared statements
		Logger:                 logger.Default.LogMode(logger.Info), // Enables logging of SQL statements
	})
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	// AutoMigrate your models
	if err := db.AutoMigrate(&model.User{}, &model.Transaction{}); err != nil {
		log.Fatalf("Error during migration: %v", err)
	}

	// Check if the default customer exists
	var defaultUser model.User
	if err := db.First(&defaultUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create the default user if not found
			defaultUser = model.User{
				Balance: 0,
			}
			if err := db.Create(&defaultUser).Error; err != nil {
				return fmt.Errorf("error creating default user: %v", err)
			}
		}
	}
	return nil
}
func (indexRepo indexRepo) Getconnected() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in routes ", err)
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_TIMEZONE"),
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,  // Connection string for Postgres
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		SkipDefaultTransaction: true,                                // Skip default transactions for performance
		PrepareStmt:            true,                                // Caches prepared statements
		Logger:                 logger.Default.LogMode(logger.Info), // Enables logging of SQL statements
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	return db, nil
}
func (indexRepo indexRepo) DbClose(GormDB *gorm.DB) {
	sqlDB, err := GormDB.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}

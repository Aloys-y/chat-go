package db

import (
	"fmt"
	"log"

	"github.com/yourusername/chat-go/config"
	"github.com/yourusername/chat-go/models"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

var DB *gorm.DB

// InitDB initializes the MySQL database connection
func InitDB() error {
	cfg := config.AppConfig.Database

	// MySQL DSN format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset, cfg.ParseTime)

	var err error
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL database: %w", err)
	}

	// Enable logging
	DB.LogMode(true)

	// Auto migrate models
	if err := DB.AutoMigrate(&models.User{}, &models.Room{}).Error; err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("MySQL database connected and migrated successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
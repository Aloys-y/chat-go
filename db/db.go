package db

import (
	"fmt"
	"log"

	"github.com/Aloys-y/chat-go/config"
	"github.com/Aloys-y/chat-go/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the MySQL database connection
func InitDB() error {
	cfg := config.AppConfig.Database

	// MySQL DSN format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset, cfg.ParseTime)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL database: %w", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database object: %w", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)

	// Auto migrate models
	if err := DB.AutoMigrate(&models.User{}, &models.Room{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("MySQL database connected and migrated successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Failed to get database object: %v", err)
			return
		}
		sqlDB.Close()
	}
}

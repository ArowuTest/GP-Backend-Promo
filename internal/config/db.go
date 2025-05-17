package config

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "log"
    "os"
    "time"
)

// DBInstance holds the database connection
type DBInstance struct {
    *gorm.DB
}

// InitDB initializes the database connection
func InitDB() (*DBInstance, error) {
    // Ensure config is initialized
    if Config.DBHost == "" {
        if err := Initialize(); err != nil {
            return nil, err
        }
    }

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
        Config.DBHost,
        Config.DBUser,
        Config.DBPassword,
        Config.DBName,
        Config.DBPort,
        Config.DBSSLMode,
    )

    // Configure GORM logger
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold: time.Second,
            LogLevel:      logger.Info,
            Colorful:      true,
        },
    )

    // Open connection
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })
    if err != nil {
        return nil, err
    }

    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return &DBInstance{db}, nil
}

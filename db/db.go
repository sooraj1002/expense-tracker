package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sooraj1002/expense-tracker/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *sql.DB

type Expense struct{
	gorm.Model
	Amount float64 `gorm:"column:amount"`
	Date string	`gorm:"column:date"`
}

func InitDB(filepath string)(*gorm.DB, error){
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	
	if err != nil {
		return nil,err
	}

	db.AutoMigrate(&Expense{})

	logger.Log.Info("Db has been initialized")
	return db, nil
}
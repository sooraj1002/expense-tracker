package main

import (
	"github.com/sooraj1002/expense-tracker/cmd"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
)

func main() {
	logger.InitLogger()
	logger.Log.Info("logger has been initialized")

	db.InitDB("./expenses.db")
	cmd.Execute()
}

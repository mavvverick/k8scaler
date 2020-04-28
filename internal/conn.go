package internal

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// ConnectSQL to connect to mysql db
func ConnectSQL() (*gorm.DB, error) {
	host := os.Getenv("SQL_HOST")
	user := os.Getenv("SQL_USER")
	pass := os.Getenv("SQL_PASS")
	port := os.Getenv("SQL_PORT")
	dbname := os.Getenv("SQL_DB")

	args := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True",
		user,
		pass,
		host,
		port,
		dbname,
	)
	db, err := gorm.Open("mysql", args)
	if err != nil {
		return nil, err
	}
	return db, nil
}

package persistence

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Database struct {
	*gorm.DB
}

var DB *gorm.DB

// Opening a database and save the reference to `Database` struct.
func Init() *gorm.DB {
	dsn := "root:root@tcp(localhost:33060)/chat?parseTime=True"
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		fmt.Println("db err: (Init) ", err)
	}

	db.AutoMigrate(&User{})

	return db
}

package persistence

import (
	"fmt"
	"go-chat/settings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Database struct {
	*gorm.DB
}

var DB *gorm.DB

// Opening a database and save the reference to `Database` struct.
func Init(cfg *settings.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?parseTime=True",
		cfg.Database.User,
		cfg.Database.Pass,
		cfg.Database.Protocol,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DataSource)

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		fmt.Println("db err: (Init) ", err)
	}

	db.AutoMigrate(&User{})

	DB = db

	return db
}

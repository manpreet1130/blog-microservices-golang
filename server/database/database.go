package database

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

func Connect() {
	d, err := gorm.Open("sqlite3", "./server/database/blog.db")
	if err != nil {
		log.Fatal("could not connect to db")
	}
	db = d
}

func GetDB() *gorm.DB {
	return db
}

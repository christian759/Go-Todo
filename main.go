package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title     string
	Completed bool
}

var db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

func main() {
	db.AutoMigrate(&Todo{})
}

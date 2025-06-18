package main

import (
	"fmt"
	"os"

	"github.com/aquasecurity/table"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title     string `gorm:"uniqueIndex:idx_title"`
	Completed bool
	Priority  string
}

func (t Todo) String() string {
	return fmt.Sprintf("Todo Title: %s, Completed: %t, Priority: %s", t.Title, t.Completed, t.Priority)
}

var db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

func createTodo(title string, priority string) Todo {
	var completed bool = false
	newTodo := Todo{Title: title, Completed: completed, Priority: priority}
	if res := db.Create(&newTodo); res.Error != nil {
		panic(res.Error)
	}
	return newTodo
}

func getTodo(title string) Todo {
	targetPost := Todo{Title: title}
	if res := db.First(&targetPost); res.Error != nil {
		panic(res.Error)
	}
	return targetPost
}

func main() {
	// db.AutoMigrate(&Todo{})
	// freshPost := createTodo("Sleep", "High")
	// oldPost := getTodo("Sleep")
	// fmt.Println(oldPost)
	t := table.New(os.Stdout)

	t.SetHeaders("ID", "Fruit", "Stock")

	t.AddRow("1", "Apple", "14")
	t.AddRow("2", "Banana", "88,041")
	t.AddRow("3", "Cherry", "342")
	t.AddRow("4", "Dragonfruit", "1")

	t.Render()
}

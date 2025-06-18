package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aquasecurity/table"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title     string
	Completed bool
	Priority  string
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	db.AutoMigrate(&Todo{})
}

// deleting after 30 days
func cleanupOldTodos() {
	db.Where("created_at < ?", time.Now().AddDate(0, -1, 0)).Delete(&Todo{})
}

// listTodos returns all todos ordered by ID ascending
func listTodos() []Todo {
	var todos []Todo
	db.Order("id ASC").Find(&todos)
	return todos
}

// createTodo tries to create a Todo with given title and priority.
// If a Todo with the same title already exists, it returns an error instead of panicking.
func createTodo(title, priority string) error {
	// Check if a todo with this title already exists (case-sensitive).
	var existing Todo
	err := db.Where("title = ?", title).First(&existing).Error
	if err == nil {
		// Found an existing todo with same title
		return fmt.Errorf("todo with title %q already exists", title)
	}
	if err != gorm.ErrRecordNotFound {
		// Some other DB error
		return fmt.Errorf("error checking existing todo: %w", err)
	}
	// Not found, proceed to create
	if err := db.Create(&Todo{Title: title, Priority: priority}).Error; err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}
	return nil
}

func updateTodo(id uint, title, priority string) error {
	var todo Todo
	if err := db.First(&todo, id).Error; err != nil {
		return fmt.Errorf("todo not found: %w", err)
	}
	todo.Title = title
	todo.Priority = priority
	if err := db.Save(&todo).Error; err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}
	return nil
}

func toggleCompleted(id uint) error {
	var todo Todo
	if err := db.First(&todo, id).Error; err != nil {
		return fmt.Errorf("todo not found: %w", err)
	}
	todo.Completed = !todo.Completed
	if err := db.Save(&todo).Error; err != nil {
		return fmt.Errorf("failed to toggle todo: %w", err)
	}
	return nil
}

func deleteTodo(id uint) error {
	if err := db.Delete(&Todo{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	return nil
}

// renderTable prints the current todos in a CLI table, ordered by ID ascending.
func renderTable() {
	t := table.New(os.Stdout)
	t.SetPadding(4)
	t.SetHeaders("Title", "Completed", "Priority", "CreatedAt")

	var todos []Todo
	db.Order("id ASC").Find(&todos)
	for _, todo := range todos {
		comp := "❌"
		if todo.Completed {
			comp = "✅"
		}
		t.AddRow(
			todo.Title,
			comp,
			todo.Priority,
			todo.CreatedAt.Format("2006-01-02"),
		)
	}
	t.Render()
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, listTodos())
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		priority := r.FormValue("priority")
		if err := createTodo(title, priority); err != nil {
			// Log the error to console. In a more advanced app, you might show a flash message in the UI.
			fmt.Fprintf(os.Stderr, "Add error: %v\n", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		title := r.FormValue("title")
		priority := r.FormValue("priority")
		if err := updateTodo(uint(id), title, priority); err != nil {
			fmt.Fprintf(os.Stderr, "Edit error: %v\n", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		if err := toggleCompleted(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Toggle error: %v\n", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		if err := deleteTodo(uint(id)); err != nil {
			fmt.Fprintf(os.Stderr, "Delete error: %v\n", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	initDB()
	cleanupOldTodos()
	renderTable()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/edit", handleEdit)
	http.HandleFunc("/toggle", handleToggle)
	http.HandleFunc("/delete", handleDelete)

	fmt.Println("🌐 Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}

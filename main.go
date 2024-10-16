package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var tmpl *template.Template
var db *sql.DB
var Store = sessions.NewCookieStore([]byte("user-management-secret"))

func init() {
	tmpl, _ = template.ParseGlob("templates/*.html")

	// Set up sessions
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 3,
		HttpOnly: true,
	}
}

func initDB() {
	// Define the PostgreSQL connection string
	connStr := "postgres://postgres:secret@localhost:5432/profile?sslmode=disable"

	// Open the connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Check if the connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
}

func main() {

	gRouter := mux.NewRouter()

	// Initialize the database
	initDB()
	defer db.Close()

	// Routes
	gRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "home.html", nil)
	})

	// Start server
	http.ListenAndServe(":4000", gRouter)
}

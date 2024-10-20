package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"htmx/pkg/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	_ "github.com/lib/pq"
)

var tmpl *template.Template
var db *sql.DB
var Store = sessions.NewCookieStore([]byte("usermanagementsecret"))

func initTemplate() {
	var err error
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

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
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Check if the connection is valid
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
}

func main() {
	gRouter := mux.NewRouter()

	// Initialize templates
	initTemplate()

	// Initialize the database
	initDB()
	defer db.Close()

	// Routes
	gRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "home.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	gRouter.HandleFunc("/register", handlers.RegisterPage(db, tmpl)).Methods("GET")

	gRouter.HandleFunc("/register", handlers.RegisterHandler(db, tmpl)).Methods("POST")

	gRouter.HandleFunc("/login", handlers.LoginPage(db, tmpl)).Methods("GET")

	gRouter.HandleFunc("/login", handlers.LoginHandler(db, tmpl, Store)).Methods("POST")

	// Start server
	http.ListenAndServe(":4000", gRouter)
}

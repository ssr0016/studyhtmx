package main

import (
	"database/sql"
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
}

func main() {
	gRouter := mux.NewRouter()

	// Initialize the template
	initTemplate()

	// Initialize the database
	initDB()
	defer db.Close()

	//File server
	fileServer := http.FileServer(http.Dir("./uploads"))
	gRouter.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", fileServer))

	// Routes
	gRouter.HandleFunc("/", handlers.Homepage(db, tmpl, Store)).Methods("GET")

	gRouter.HandleFunc("/register", handlers.RegisterPage(db, tmpl)).Methods("GET")

	gRouter.HandleFunc("/register", handlers.RegisterHandler(db, tmpl)).Methods("POST")

	gRouter.HandleFunc("/login", handlers.LoginPage(db, tmpl)).Methods("GET")

	gRouter.HandleFunc("/login", handlers.LoginHandler(db, tmpl, Store)).Methods("POST")

	gRouter.HandleFunc("/edit", handlers.Editpage(db, tmpl, Store)).Methods("GET")

	gRouter.HandleFunc("/edit", handlers.UpdateProfileHandler(db, tmpl, Store)).Methods("POST")

	gRouter.HandleFunc("/upload-avatar", handlers.AvatarPage(db, tmpl, Store)).Methods("GET")

	gRouter.HandleFunc("/upload-avatar", handlers.UploadAvatarHandler(db, tmpl, Store)).Methods("POST")

	// Start server
	http.ListenAndServe(":4000", gRouter)
}

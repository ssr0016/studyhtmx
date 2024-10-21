package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"htmx/pkg/models"
	"htmx/pkg/repository"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func RegisterPage(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "register", nil)
	}
}

func RegisterHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		var errorMessages []string

		// Parse the form data
		r.ParseForm()

		user.Name = r.FormValue("name")
		user.Email = r.FormValue("email")
		user.Password = r.FormValue("password")
		user.Category, _ = strconv.Atoi(r.FormValue("category"))

		// Basic Validation
		if user.Name == "" {
			errorMessages = append(errorMessages, "Name is required")
		}

		if user.Email == "" {
			errorMessages = append(errorMessages, "Email is required")
		}

		if user.Password == "" {
			errorMessages = append(errorMessages, "Password is required")
		}

		if len(errorMessages) > 0 {
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			errorMessages = append(errorMessages, "Failed to hash password")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}
		user.Password = string(hashedPassword)

		// Set default values
		user.DOB = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		user.Bio = "Bio goes here"
		user.Avatar = ""

		// Create user in the database
		err = repository.CreateUser(db, user)
		if err != nil {
			errorMessages = append(errorMessages, "Failed to create user: "+err.Error())
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Set HTTP status code to 204 (not content) and set 'HX-Location' header to signal HTMX to redirect
		w.Header().Set("HX-Location", "/login")
		w.WriteHeader(http.StatusNoContent)
	}
}

func LoginPage(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "login", nil)
	}
}

func LoginHandler(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		email := r.FormValue("email")
		password := r.FormValue("password")

		var errorMessages []string

		// Basic Validation
		if email == "" {
			errorMessages = append(errorMessages, "Email is required")
		}

		if password == "" {
			errorMessages = append(errorMessages, "Password is required")
		}

		if len(errorMessages) > 0 {
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Retrieve user by email
		user, err := repository.GetUserByEmail(db, email)
		if err != nil {
			if err == sql.ErrNoRows {
				errorMessages = append(errorMessages, "Invalid email or password")
				tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Compare passwords
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			errorMessages = append(errorMessages, "Invalid email or password")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Create session and authenticate the user
		session, err := store.Get(r, "logged-in-user")
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		session.Values["user_id"] = user.Id
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Error saving session", http.StatusInternalServerError)
			return
		}

		// Set HX-Location header and return 204 No Content Status
		w.Header().Set("HX-Location", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func Homepage(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := CheckLoggedIn(w, r, store, db)
		if err != nil {
			return // Error already handled in `CheckLoggedIn`
		}

		if err := tmpl.ExecuteTemplate(w, "home.html", user); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func Editpage(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := CheckLoggedIn(w, r, store, db)
		if err != nil {
			return // Error already handled in `CheckLoggedIn`
		}

		if err := tmpl.ExecuteTemplate(w, "editProfile", user); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	}
}

func UpdateProfileHandler(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve user from session
		currentUserProfile, userIDErr := CheckLoggedIn(w, r, store, db)
		if userIDErr != nil {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		// Ensure userID is the correct type (string)
		userID := currentUserProfile.Id // Assuming `Id` is of type string in your user model

		// Parse the form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusInternalServerError)
			return
		}

		var errorMessages []string

		// Collect and validate form data
		name := r.FormValue("name")
		bio := r.FormValue("bio")
		dobStr := r.FormValue("dob")

		if name == "" {
			errorMessages = append(errorMessages, "Name is required")
		}

		if dobStr == "" {
			errorMessages = append(errorMessages, "Date of Birth is required")
		}

		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil {
			errorMessages = append(errorMessages, "Invalid date format")
		}

		// Handle validation errors
		if len(errorMessages) > 0 {
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Create user struct
		user := models.User{
			Id:     userID, // Ensure correct assignment
			Name:   name,
			DOB:    dob,
			Bio:    bio,
			Avatar: currentUserProfile.Avatar,
		}

		// Call the repository function to update the user
		if err := repository.UpdateUser(db, userID, user); err != nil {
			errorMessages = append(errorMessages, "Failed to update user")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			log.Fatal(err) // Consider logging instead of using log.Fatal in production
			return
		}

		// Redirect or return success
		w.Header().Set("HX-Location", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func AvatarPage(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := CheckLoggedIn(w, r, store, db)
		if err != nil {
			return // Error already handled in `CheckLoggedIn`
		}

		if err := tmpl.ExecuteTemplate(w, "uploadAvatar", user); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func UploadAvatarHandler(db *sql.DB, tmpl *template.Template, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, userIDErr := CheckLoggedIn(w, r, store, db)
		if userIDErr != nil {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		// Ensure userID is the correct type (string)
		userID := user.Id

		var errorMessages []string

		// Parse the multipart form, 10 MB max upload size
		r.ParseMultipartForm(10 << 20)

		// Retrieve the file from form data
		file, handler, err := r.FormFile("avatar")
		if err != nil {
			if err == http.ErrMissingFile {
				errorMessages = append(errorMessages, "No file Submitted")
			} else {
				errorMessages = append(errorMessages, "Error retrieving the file")
			}

			if len(errorMessages) > 0 {
				tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
				return
			}
		}
		defer file.Close()

		// Generate a unique filename to prevent overwriting and conflicts
		uuid, err := uuid.NewRandom()
		if err != nil {
			errorMessages = append(errorMessages, "Error generating UUID identifier")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}
		// Append the file extension
		filename := uuid.String() + filepath.Ext(handler.Filename)

		// Create the full path for saving the file
		filePath := filepath.Join("uploads", filename)

		// Save the file to the server
		dst, err := os.Create(filePath)
		if err != nil {
			errorMessages = append(errorMessages, "Error saving the file")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			errorMessages = append(errorMessages, "Error saving the file")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

		// Update the user's avatar in the database
		if err := repository.UpdateUserAvatar(db, userID, filename); err != nil {
			errorMessages = append(errorMessages, "Error updating user avatar")
			tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			log.Fatal(err)
			return
		}

		// Delete current image from the initial fetch of the user
		if user.Avatar != "" {
			oldAvatarPath := filepath.Join("uploads", user.Avatar)

			// Check if the oldPath is not the same as the newPath
			if oldAvatarPath != filePath {
				if err := os.Remove(oldAvatarPath); err != nil {
					log.Printf("Warning: failed to delete old avatar file: %s\n", err)
				}
			}
		}

		// Navigate to the profile page after the update
		w.Header().Set("HX-Location", "/")
		w.WriteHeader(http.StatusNoContent)
	}
}

func CheckLoggedIn(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB) (models.User, error) {
	session, err := store.Get(r, "logged-in-user")
	if err != nil {
		log.Println("Session retrieval error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return models.User{}, err
	}

	userID, ok := session.Values["user_id"]
	if !ok {
		log.Println("User ID not found in session, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return models.User{}, fmt.Errorf("user not logged in")
	}

	user, err := repository.GetUserById(db, userID.(string))
	if err != nil {
		if err == sql.ErrNoRows {
			session.Options.MaxAge = -1 // Clear session
			session.Save(r, w)
			log.Println("No user found, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return models.User{}, fmt.Errorf("user not found")
		}
		log.Println("Database error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return models.User{}, err
	}

	return user, nil
}

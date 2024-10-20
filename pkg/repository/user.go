package repository

import (
	"database/sql"
	"errors"
	"htmx/pkg/models"

	"github.com/google/uuid"
)

func GetAllUsers(db *sql.DB) ([]models.User, error) {
	users := []models.User{}

	query := `
		SELECT
			id,
			email,
			password,
			name,
			category,
			dob,
			bio,
			avatar
		FROM users
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		user := models.User{}
		err = rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Category,
			&user.DOB,
			&user.Bio,
			&user.Avatar,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserById(db *sql.DB, userID string) (models.User, error) {
	var user models.User

	query := `
		SELECT
	 		id,
			name,
			email,
			avatar,
			bio,
			category,
			dob 
		FROM
		 	users
		WHERE
		 	id = $1`
	err := db.QueryRow(query, userID).Scan(&user.Id, &user.Name, &user.Email, &user.Avatar, &user.Bio, &user.Category, &user.DOB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, sql.ErrNoRows
		}
		return user, err
	}

	user.DOBFormatted = user.DOB.Format("2006-01-02")

	return user, nil
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User

	// Use PostgreSQL-style placeholder ($1)
	err := db.QueryRow(`
		SELECT
			id,
			email,
			password,
			name,
			category,
			dob,
			bio,
			avatar
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Category,
		&user.DOB,
		&user.Bio,
		&user.Avatar,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return a more specific error if no user is found
			return user, nil
		}
		return user, err
	}

	return user, nil
}

func CreateUser(db *sql.DB, user models.User) error {
	// Generate a new UUID for the user
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	// Set the generated ID on the user object
	user.Id = id.String()

	// Use PostgreSQL-style placeholders ($1, $2, ...)
	stmt, err := db.Prepare(`
		INSERT INTO users
		(id, email, password, name, category, dob, bio, avatar)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close() // Ensure the statement is closed properly

	// Execute the statement with user data
	_, err = stmt.Exec(
		user.Id,
		user.Email,
		user.Password,
		user.Name,
		user.Category,
		user.DOB,
		user.Bio,
		user.Avatar,
	)

	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(db *sql.DB, id string, user models.User) error {
	query := `
		UPDATE users
		SET
			name = $1,
			category = $2,
			dob = $3,
			bio = $4
		WHERE
			id = $5
	`
	_, err := db.Exec(
		query,
		user.Name,
		user.Category,
		user.DOB,
		user.Bio,
		id,
	)

	return err
}

func UpdateUserAvatar(db *sql.DB, userID, filename string) error {
	query := `
		UPDATE users
		SET
			avatar = $1
		WHERE
			id = $2
	`

	_, err := db.Exec(
		query,
		filename,
		userID,
	)

	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(db *sql.DB, id string) error {
	_, err := db.Exec(
		`DELETE
		 FROM
		 	users
		 WHERE
		 	id = ?
		`,
		id,
	)

	return err
}

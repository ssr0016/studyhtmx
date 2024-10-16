package repository

import (
	"database/sql"
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

func GetUserById(db *sql.DB, id string) (models.User, error) {
	var user models.User

	err := db.QueryRow(
		`SELECT
		 	id,
			email,
			password,
			name,
			category,
			dob,
			bio,
			avatar
		FROM users
		WHERE
			id = ?	
		`,
		id,
	).Scan(
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
		return user, err
	}

	// Format the date using a friendly format
	user.DOBFormatted = user.DOB.Format("2006-01-02")

	return user, nil
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User

	err := db.QueryRow(
		`SELECT
		 	id,
			email,
			password,
			name,
			category,
			dob,
			bio,
			avatar
		FROM users
		WHERE
			email = ?	
		`,
		email,
	).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Category,
		&user.DOB,
		&user.Bio,
		&user.Avatar,
	)

	return user, err
}

func CreateUser(db *sql.DB, user models.User) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	// Convert id to string and set it on the user
	user.Id = id.String()

	stmt, err := db.Prepare(`
		INSERT INTO users
		(id, email, password, name, category, dob, bio, avatar)
		ValUES
		(?, ?, ?, ?, ?, ?, ?, ?)
	`)

	if err != nil {
		return err
	}

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

func UpdateUser(db *sql.DB, user models.User) error {
	_, err := db.Exec(
		`UPDATE users
		 SET
		 	name = ?,
		 	category = ?,
		 	dob = ?,
		 	bio = ?
		 WHERE
		 	id = ?
		`,
		user.Name,
		user.Category,
		user.DOB,
		user.Bio,
		user.Id,
	)

	return err
}

func UpdateUserAvatar(db *sql.DB, userID, filePath string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET
		 	avatar = ?
		 WHERE
		 	id = ?
		`,
		filePath,
		userID,
	)

	return err
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

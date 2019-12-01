package users

import (
	"database/sql"
	"fmt"
	"time"
)

// baseSelectStatement is SQL select statement that retrieves all user data from the Users table
// This base select statement is reused many times in this file thus justifying it's existence
// as a global constant
const baseSelectStatement = "SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users "

// MySQLStore represents a users.Store backed by MySQL.
type MySQLStore struct {
	Client *sql.DB
}

// NewMySQLStore constructs a new MySQLStore
func NewMySQLStore(client *sql.DB) *MySQLStore {
	return &MySQLStore{client}
}

// GetByID returns the User with the given ID
func (ms *MySQLStore) GetByID(id int64) (*User, error) {
	selectQuery := baseSelectStatement + "WHERE ID = ?"
	return getUser(ms.Client, selectQuery, id)
}

// GetByEmail returns the User with the given email
func (ms *MySQLStore) GetByEmail(email string) (*User, error) {
	selectQuery := baseSelectStatement + "WHERE Email = ?"
	return getUser(ms.Client, selectQuery, email)
}

// GetByUserName returns the User with the given Username
func (ms *MySQLStore) GetByUserName(username string) (*User, error) {
	selectQuery := baseSelectStatement + "WHERE UserName = ?"
	return getUser(ms.Client, selectQuery, username)
}

// Insert inserts the user into the database, and returns
// the newly-inserted User, complete with the DBMS-assigned ID
func (ms *MySQLStore) Insert(user *User) (*User, error) {
	insertQuery := "INSERT INTO Users(Email, PassHash, UserName, FirstName, LastName, PhotoURL) VALUES(?,?,?,?,?,?)"

	result, err := ms.Client.Exec(insertQuery, user.Email, user.PassHash, user.UserName,
		user.FirstName, user.LastName, user.PhotoURL)	
	if err != nil {
		return user, fmt.Errorf("Error inserting new user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return user, fmt.Errorf("Error getting new user ID: %v", err)
	}

	user.ID = id	
	return user, nil
}

// Update applies UserUpdates to the given user ID and returns the newly-updated user
func (ms *MySQLStore) Update(id int64, updates *Updates) (*User, error) {
	updateQuery := "UPDATE Users SET FirstName = ?, LastName = ? WHERE ID = ?"

	_, err := ms.Client.Exec(updateQuery, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, fmt.Errorf("Error updating user: %v", err)
	}

	selectQuery := baseSelectStatement + "WHERE ID = ?"
	return getUser(ms.Client, selectQuery, id)
}

// Delete deletes the user with the given ID
func (ms *MySQLStore) Delete(id int64) error {
	deletionQuery := "DELETE FROM Users WHERE ID = ?"

	_, err := ms.Client.Exec(deletionQuery, id)
	if err != nil {
		return fmt.Errorf("Error deleting user: %v", err)
	}

	return nil
}

// LogUser logs a successful sign-in by a user with the user ID, curent time,
// and user IP address
func (ms *MySQLStore) LogUser(id int64, time time.Time, clientIP string) error {
	insertQuery := "INSERT INTO UserSignInLog(UserID, SignInTime, ClientIP) VALUES(?,?,?)"

	_, err := ms.Client.Exec(insertQuery, id, time, clientIP)
	if err != nil {
		return fmt.Errorf("Error logging user sign in: %v", err)
	}

	return nil
}

// getUser is a helper function for getting a specific user based on a given SQL select statement
// and select parameter.
// Note: selectParam has the type: interface{}, meaning a variable with any type can be passed
// into the function, thus allowing both int64 (id) and string (email, username) to be passed to
// the same helper function.
func getUser(db *sql.DB, selectQuery string, selectParam interface{}) (*User, error) {
	user := &User{}

	rows, err := db.Query(selectQuery, selectParam)
	if err != nil {
		return user, fmt.Errorf("Error selecting user: %v", err)
	}

	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL)
		if err != nil {
			return user, fmt.Errorf("Error scanning selected user: %v", err)
		}
	}

	if err := rows.Err(); err != nil {
		return user, fmt.Errorf("Error fetching selected user: %v", err)
	}

	defer rows.Close()

	if user.UserName == "" {
		return user, fmt.Errorf("User not found")
	}
	return user, nil
}

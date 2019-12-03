package users

import (
	"fmt"
	"time"
)

// TestUserStore represents a UserStore with mock data
type TestUserStore struct {
	Client string
}

// NewTestUserStore creates a NewTestUserStore
func NewTestUserStore(client string) *TestUserStore {
	return &TestUserStore{client}
}

// GetByID returns the User with the given ID
func (client *TestUserStore) GetByID(id int64) (*User, error) {
	if id == 1 {
		newUser := &NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
		user, _ := newUser.ToUser()
		user.ID = 1
		return user, nil
	} else {
		return nil, fmt.Errorf("Error fetching selected user")
	}
}

// GetByEmail returns the User with the given email
func (client *TestUserStore) GetByEmail(email string) (*User, error) {
	if email == "stanley@gmail.com" {
		newUser := &NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
		user, _ := newUser.ToUser()
		return user, nil
	}
	return nil, ErrUserNotFound
}

// GetByUserName returns the User with the given Username
func (client *TestUserStore) GetByUserName(username string) (*User, error) {
	return nil, nil
}

// Insert inserts the user into the database, and returns
// the newly-inserted User, complete with the DBMS-assigned ID
func (client *TestUserStore) Insert(user *User) (*User, error) {
	newUser := &NewUser{Email: "stanley@gmail.com", Password: "123456", PasswordConf: "123456", UserName: "swu", FirstName: "Stanley", LastName: "Wu"}
	createdUser, _ := newUser.ToUser()
	createdUser.ID = 1
	return createdUser, nil
}

// Update applies UserUpdates to the given user ID and returns the newly-updated user
func (client *TestUserStore) Update(id int64, updates *Updates) (*User, error) {
	return nil, nil
}

// Delete deletes the user with the given ID
func (client *TestUserStore) Delete(id int64) error {
	return nil
}

// LogUser logs a successful sign-in by a user with the user ID, curent time,
// and user IP address
func (client *TestUserStore) LogUser(id int64, time time.Time, clientIP string) error {
	return nil
}

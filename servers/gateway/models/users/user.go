package users

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// gravatarBasePhotoURL is the base URL for Gravatar image requests.
// See https://id.gravatar.com/site/implement/images/ for details
const gravatarBasePhotoURL = "https://www.gravatar.com/avatar/"

// bcryptCost is the default bcrypt cost to use when hashing passwords
var bcryptCost = 13

// User represents a user account in the database
type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"-"` //never JSON encoded/decoded
	PassHash  []byte `json:"-"` //never JSON encoded/decoded
	UserName  string `json:"userName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	PhotoURL  string `json:"photoURL"`
}

// Credentials represents user sign-in credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewUser represents a new user signing up for an account
type NewUser struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	PasswordConf string `json:"passwordConf"`
	UserName     string `json:"userName"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
}

// Updates represents allowed updates to a user profile
type Updates struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// Validate validates the new user and returns an error if
// any of the validation rules fail, or nil if its valid
func (nu *NewUser) Validate() error {
	_, err := mail.ParseAddress(nu.Email)
	if err != nil {
		return fmt.Errorf("Email not valid")
	}
	if len(nu.Password) < 6 {
		return fmt.Errorf("Password has fewer than 6 characters")
	}
	if nu.Password != nu.PasswordConf {
		return fmt.Errorf("Password and password confirmation do not match")
	}
	if len(nu.UserName) == 0 || strings.Contains(nu.UserName, " ") {
		return fmt.Errorf("Username must be greater than 0 length and cannot contain spaces")
	}

	return nil
}

// ToUser converts the NewUser to a User, setting the
// PhotoURL and PassHash fields appropriately
func (nu *NewUser) ToUser() (*User, error) {
	validate := nu.Validate()
	if validate != nil {
		user := &User{}
		return user, validate
	}

	noWhiteSpaceEmail := strings.TrimSpace(nu.Email)
	lowerCaseEmail := strings.ToLower(noWhiteSpaceEmail)
	hasher := md5.New()
	hasher.Write([]byte(lowerCaseEmail))
	photoURL := "https://www.gravatar.com/avatar/" + hex.EncodeToString(hasher.Sum(nil))

	user := &User{
		Email:     nu.Email,
		UserName:  nu.UserName,
		FirstName: nu.FirstName,
		LastName:  nu.LastName,
		PhotoURL:  photoURL,
	}

	user.SetPassword(nu.Password)
	return user, nil
}

// FullName returns the user's full name, in the form:
// 	 "<FirstName> <LastName>"
// If either first or last name is an empty string, no
// space is put between the names. If both are missing,
// this returns an empty string
func (u *User) FullName() string {
	if u.FirstName == "" {
		return u.LastName
	} else if u.LastName == "" {
		return u.FirstName
	} else if u.FirstName == "" && u.LastName == "" {
		return ""
	} else {
		return u.FirstName + " " + u.LastName
	}
}

// SetPassword hashes the password and stores it in the PassHash field
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		fmt.Printf("error generating bcrypt hash: %v\n", err)
		return err
	}

	u.PassHash = hash
	return nil
}

// Authenticate compares the plaintext password against the stored hash
// and returns an error if they don't match, or nil if they do
func (u *User) Authenticate(password string) error {
	if err := bcrypt.CompareHashAndPassword(u.PassHash, []byte(password)); err != nil {
		return err
	}
	return nil
}

// ApplyUpdates applies the updates to the user. An error
// is returned if the updates are invalid
func (u *User) ApplyUpdates(updates *Updates) error {
	if updates.FirstName != "" || updates.LastName != "" {
		if updates.FirstName != "" {
			u.FirstName = updates.FirstName
		}
		if updates.LastName != "" {
			u.LastName = updates.LastName
		}
		return nil
	}
	return fmt.Errorf("Invalid update")
}

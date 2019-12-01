package users

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		input          NewUser
		expectedOutput string
	}{
		{NewUser{"badEmail", "123456", "123456", "stanley", "Stanley", "Wu"}, "Email not valid"},
		{NewUser{"123@gmail.com", "12345", "123456", "stanley", "Stanley", "Wu"}, "Password has fewer than 6 characters"},
		{NewUser{"123@gmail.com", "123456", "1234567", "stanley", "Stanley", "Wu"}, "Password and password confirmation do not match"},
		{NewUser{"123@gmail.com", "123456", "123456", "", "Stanley", "Wu"}, "Username must be greater than 0 length and cannot contain spaces"},
		{NewUser{"123@gmail.com", "123456", "123456", "stanley ", "Stanley", "Wu"}, "Username must be greater than 0 length and cannot contain spaces"},
	}

	for _, c := range cases {
		newUser := c.input
		if output := newUser.Validate(); output.Error() != c.expectedOutput {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.input, c.expectedOutput, output)
		}
	}

	newUser := NewUser{"123@gmail.com", "123456", "123456", "stanley", "Stanley", "Wu"}
	if newUser.Validate() != nil {
		t.Errorf("incorrect output for a valid user")
	}
}

func TestToUser(t *testing.T) {
	badNewUser := NewUser{"n", "My secure password000000", "My secure password000000", "stanley", "Stanley", "Wu"}
	_, err := badNewUser.ToUser()
	if err.Error() != "Email not valid" {
		t.Errorf("incorrect output invalid user")
	}

	newUser := NewUser{"Swsdfiooi@gmail.com", "My secure password", "My secure password", "stanley", "Stanley", "Wu"}
	user, err := newUser.ToUser()
	if err != nil {
		t.Errorf("incorrect output for valid user")
	}
	noWhiteSpaceEmail := strings.TrimSpace(newUser.Email)
	lowerCaseEmail := strings.ToLower(noWhiteSpaceEmail)
	hasher := md5.New()
	hasher.Write([]byte(lowerCaseEmail))
	photoURL := "https://www.gravatar.com/avatar/" + hex.EncodeToString(hasher.Sum(nil))

	cases := []struct {
		input          string
		expectedOutput string
	}{
		{user.Email, "Swsdfiooi@gmail.com"},
		{user.UserName, "stanley"},
		{user.FirstName, "Stanley"},
		{user.LastName, "Wu"},
		{user.PhotoURL, photoURL},
	}
	for _, c := range cases {
		if c.input != c.expectedOutput {
			t.Errorf("incorrect output for `%s`: expected `%s`", c.input, c.expectedOutput)
		}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 13)
	if err := bcrypt.CompareHashAndPassword(hash, []byte(newUser.Password)); err != nil {
		t.Errorf("incorrect output for password hash: '%s'", err)
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		input          User
		expectedOutput string
	}{
		{User{FirstName: "John"}, "John"},
		{User{LastName: "Smith"}, "Smith"},
		{User{FirstName: "John", LastName: "Smith"}, "John Smith"},
		{User{}, ""},
	}

	for _, c := range cases {
		user := c.input
		if output := user.FullName(); output != c.expectedOutput {
			t.Errorf("incorrect output for FullName function")
		}
	}
}

func TestAuthenticate(t *testing.T) {
	newUser := NewUser{"Swsdfiooi@gmail.com", "123456", "123456", "stanley", "Stanley", "Wu"}
	user, _ := newUser.ToUser()

	cases := []struct {
		input          string
		expectedOutput error
	}{
		{"123", bcrypt.CompareHashAndPassword(user.PassHash, []byte(newUser.Password))},
		{"", bcrypt.CompareHashAndPassword(user.PassHash, []byte(newUser.Password))},
	}
	for _, c := range cases {
		if output := user.Authenticate(c.input); output == nil {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.input, c.expectedOutput, output)
		}
	}

	if output := user.Authenticate(newUser.Password); output != nil {
		t.Errorf("incorrect output for correct password: `%s`", output)
	}
}

func TestApplyUpdates(t *testing.T) {
	user := User{FirstName: "John", LastName: "Smith"}
	updates := &Updates{FirstName: "Stan", LastName: "Lee"}
	err := user.ApplyUpdates(updates)
	if err == nil && user.FirstName != "Stan" && user.LastName != "Lee" {
		t.Errorf("User incorrectly updated")
	}

	badUpdates := &Updates{}
	errTwo := user.ApplyUpdates(badUpdates)
	if errTwo.Error() != "Invalid update" {
		t.Error("Invalid update error is not caught")
	}
}

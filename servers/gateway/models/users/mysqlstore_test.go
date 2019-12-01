package users

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	columns := []string{"ID", "Email", "PassHash", "UserName", "FirstName", "LastName", "PhotoURL"}
	mySQLStore := NewMySQLStore(db)

	mock.ExpectQuery("SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(user.ID, user.Email, user.PassHash,
			user.UserName, user.FirstName, user.LastName, user.PhotoURL))

	_, funcErr := mySQLStore.GetByID(user.ID)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	mock.ExpectQuery("SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users").
		WithArgs(3).
		WillReturnError(fmt.Errorf("Error selecting user"))

	_, funcErr2 := mySQLStore.GetByID(3)
	if funcErr2 == nil {
		t.Error("Expected error, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	columns := []string{"ID", "Email", "PassHash", "UserName", "FirstName", "LastName", "PhotoURL"}
	mySQLStore := NewMySQLStore(db)

	mock.ExpectQuery("SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users").
		WithArgs("hawkticehurst@gmail.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(user.ID, user.Email, user.PassHash,
			user.UserName, user.FirstName, user.LastName, user.PhotoURL))

	_, funcErr := mySQLStore.GetByEmail(user.Email)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetByUserName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	columns := []string{"ID", "Email", "PassHash", "UserName", "FirstName", "LastName", "PhotoURL"}
	mySQLStore := NewMySQLStore(db)

	mock.ExpectQuery("SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users").
		WithArgs("hawkticehurst").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(user.ID, user.Email, user.PassHash,
			user.UserName, user.FirstName, user.LastName, user.PhotoURL))

	_, funcErr := mySQLStore.GetByUserName(user.UserName)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	mySQLStore := NewMySQLStore(db)

	mock.ExpectExec("INSERT INTO Users").
		WithArgs(user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, funcErr := mySQLStore.Insert(user)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	emptyUser, err := generateEmptyUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	mock.ExpectExec("INSERT INTO Users").
		WithArgs(emptyUser.Email, emptyUser.PassHash, emptyUser.UserName, emptyUser.FirstName,
			emptyUser.LastName, emptyUser.PhotoURL).
		WillReturnError(fmt.Errorf("Error inserting new user"))

	_, funcErr2 := mySQLStore.Insert(emptyUser)
	if funcErr2 == nil {
		t.Error("Expected error, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	updates, err := generateUserUpdates()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the user update struct", err)
	}

	columns := []string{"ID", "Email", "PassHash", "UserName", "FirstName", "LastName", "PhotoURL"}
	mySQLStore := NewMySQLStore(db)

	mock.ExpectExec("UPDATE Users SET FirstName").
		WithArgs(updates.FirstName, updates.LastName, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL FROM Users").
		WithArgs(user.ID).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(user.ID, user.Email, user.PassHash,
			user.UserName, user.FirstName, user.LastName, user.PhotoURL))

	_, funcErr := mySQLStore.Update(user.ID, updates)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	mock.ExpectExec("UPDATE Users SET FirstName").
		WithArgs(updates.FirstName, updates.LastName, 3).
		WillReturnError(fmt.Errorf("Error updating user"))

	_, funcErr2 := mySQLStore.Update(3, updates)
	if funcErr2 == nil {
		t.Error("Expected error, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	user, err := generateBasicUser()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when generating the test user struct", err)
	}

	mySQLStore := NewMySQLStore(db)

	mock.ExpectExec("DELETE FROM Users").
		WithArgs(user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	funcErr := mySQLStore.Delete(user.ID)
	if funcErr != nil {
		t.Errorf("Expected no error, but got %v instead", err)
	}

	mock.ExpectExec("DELETE FROM Users").
		WithArgs(3).
		WillReturnError(fmt.Errorf("Error deleting user"))

	funcErr2 := mySQLStore.Delete(3)
	if funcErr2 == nil {
		t.Error("Expected error, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

// generateBasicUser Helper function for generating a basic user struct to be used in testing
func generateBasicUser() (*User, error) {
	// Generate a password hash
	pwdhash, err := generatePwdHash()
	if err != nil {
		return nil, fmt.Errorf("an error '%s' was not expected when generating a password hash", err)
	}

	// Return basic user struct
	return &User{
		1,
		"hawkticehurst@gmail.com",
		pwdhash,
		"hawkticehurst",
		"Hawk",
		"Ticehurst",
		"photo.com",
	}, nil
}

// generateEmptyUser Helper function for generating a empty user struct to be used in testing
func generateEmptyUser() (*User, error) {
	// Return empty user struct
	return &User{
		0,
		"",
		[]byte{},
		"",
		"",
		"",
		"",
	}, nil
}

// generateBasicUser Helper function for generating a basic updates struct to be used in testing
func generateUserUpdates() (*Updates, error) {
	// Return updates struct
	return &Updates{
		"Stanley",
		"Wu",
	}, nil
}

// generatePwdHash Helper function for generating a password hash to be used as fake User data
func generatePwdHash() ([]byte, error) {
	pwdhash, err := bcrypt.GenerateFromPassword([]byte("password"), 0)
	if err != nil {
		return nil, fmt.Errorf("Error generating bcrypt hash: %v", err)
	}
	return pwdhash, nil
}

// createDBConnection Helper function that will create and return a reference to
// a test MySQL docker container to be used for running tests
func createDBConnection() (*sql.DB, error) {
	// Execute bash script that will build and create a test
	// MySQL docker container
	cmd := exec.Command("/bin/sh", "createdb")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Error executing creatdb bash script: %v", err)
	}
	// Create the data source name, which identifies the
	// user, password, server address, and default database
	// dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/testdb", os.Getenv("MYSQL_ROOT_PASSWORD"))
	dsn := "root:testpwd@tcp(127.0.0.1:3306)/testdb"
	log.Println(dsn)

	// Create a database object, which manages a pool of
	// network connections to the database server
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v", err)
	}

	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(151)

	// For now, just ping the server to ensure we have
	// a live connection to it
	if err := db.Ping(); err != nil {
		fmt.Printf("Error pinging database: %v\n", err)
	} else {
		fmt.Printf("Successfully connected!\n")
	}

	return db, nil
}

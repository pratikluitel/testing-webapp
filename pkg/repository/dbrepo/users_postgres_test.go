//go:build integration

package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"webapp/pkg/data"
	"webapp/pkg/repository"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {
	// connect to docker; fail if docker not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; is it running? %s", err)
	}
	pool = p

	// set up out docker options, specifying the image, etc
	opt := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	// get a resource (docker image)
	resource, err = pool.RunWithOptions(&opt)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// start the image, and wait until the database is ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))

		if err != nil {
			log.Println("Error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to database: %s", err)
	}

	// populate the database with empty tables
	err = createTables()
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	testRepo = &PostgresDBRepo{DB: testDB}

	// run tests

	code := m.Run()

	// clean up resources
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		return err
	}

	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("Can't ping the database")
	}
}

func Test_PostgresDBRepo_InsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	id, err := testRepo.InsertUser(testUser)

	if err != nil {
		t.Errorf("Insert user returned an error: %s", err)
	}

	if id != 1 {
		t.Errorf("Insert user returned wrong id; expected 1, but got %d", id)
	}
}

func Test_PostgresDBRepo_AllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()

	if err != nil {
		t.Errorf("All users returned an error: %s", err)
	}

	if len(users) != 1 {
		t.Errorf("Count of users did not match: expected 1, but got %d", len(users))
	}

	// insert another user and test count
	testUser := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "smith@example.com",
		Password:  "sesdfacret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("All users returned an error: %s", err)
	}

	if len(users) != 2 {
		t.Errorf("Count of users did not match: expected 2, but got %d", len(users))

	}
}

func Test_PostgresDBRepo_GetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("Get user by id returned an error: %s", err)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("wrong email returned by GetUser; expected admin@example.com, returned %s", user.Email)
	}

	// non existent id
	_, err = testRepo.GetUser(3)
	if err == nil {
		t.Error("no error reported when getting non existent user by id")
	}
}

func Test_PostgresDBRepo_GetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("smith@example.com")
	if err != nil {
		t.Errorf("Get user by id returned an error: %s", err)
	}
	if user.ID != 2 {
		t.Errorf("wrong id returned by GetUser; expected 2, returned %d", user.ID)
	}
	// non existent email
	_, err = testRepo.GetUserByEmail("dsfaf")
	if err == nil {
		t.Error("no error reported when getting non existent user by email")
	}
}

func Test_PostgresDBRepo_UpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(2)
	user.FirstName = "Jane"
	user.Email = "jane@example.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("error updating user %d: %s", 2, err)
	}

	user, _ = testRepo.GetUser(2)
	if user.FirstName != "Jane" || user.Email != "jane@example.com" {
		t.Errorf("Record not updated in database. Expected firstname Jane, email jane@example.com, but got %s, and %s", user.FirstName, user.Email)
	}
}

func Test_PostgresDBRepo_DeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("error deleting user %d: %s", 2, err)
	}

	_, err = testRepo.GetUser(2)

	if err == nil {
		t.Errorf("Error: expected deleted user with id %d, but was not deleted", 2)
	}
}

func Test_PostgresDBRepo_ResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "password")
	if err != nil {
		t.Error("Error resetting password", err)
	}
	user, _ := testRepo.GetUser(1)
	matches, err := user.PasswordMatches("password")
	if err != nil {
		t.Error("Error resetting password", err)
	}
	if !matches {
		t.Errorf("Password should be `password` but is not")
	}
}

func Test_PostgresDBRepo_InsertUserImage(t *testing.T) {
	var userImage = data.UserImage{
		ID:        1,
		UserID:    1,
		FileName:  "test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newID, err := testRepo.InsertUserImage(userImage)
	if err != nil {
		t.Error("Inserting user image failed:", err)
	}

	if newID != 1 { // since this is the first entry inserted, id must be 1
		t.Error("Got wrong id for image, expected 1, but got", newID)
	}

	userImage.UserID = 100 // for a user that doesn't exist
	_, err = testRepo.InsertUserImage(userImage)
	if err == nil {
		t.Error("Expected an error while inserting a user image with non existent user id, found no error")
	}
}

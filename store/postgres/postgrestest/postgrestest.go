package postgrestest

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/friendsofgo/errors"

	// import postgresq driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	env = "SFB_DATABASE"
)

func init() {
	dsn := os.Getenv(env)

	if dsn == "" {
		dsn = "postgres://sfb:sfb@localhost/sfb?sslmode=disable"
	}

	newDSN, err := createTestDB(dsn)
	if err != nil {
		fmt.Printf("create test db: %v", err)
		os.Exit(1)
	}

	txdb.Register("txdb", "postgres", newDSN)
}

func createTestDB(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", errors.Wrap(err, "parse url")
	}

	testDatabaseName := strings.TrimPrefix(u.Path, "/") + "_test"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return "", errors.Wrap(err, "open db")
	}
	defer db.Close()

	var exist bool

	if err := db.QueryRow(
		"select true from pg_database where datname = $1",
		testDatabaseName,
	).Scan(&exist); err == sql.ErrNoRows {
		exist = false
	} else if err != nil {
		return "", errors.Wrap(err, "check if db exists")
	}

	if exist {
		u.Path += "_test"
		return u.String(), nil
	}

	query := fmt.Sprintf("create database %s", testDatabaseName)
	_, err = db.Exec(query)
	if err != nil {
		return "", errors.Wrap(err, "create db")
	}

	u.Path += "_test"

	return u.String(), nil
}

// New creates a sql.DB with rollback support.
func New(t *testing.T) *sql.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skip store test because short mode")
	}

	db, err := sql.Open("txdb", t.Name())
	if err != nil {
		t.Fatalf("can't open db: %s", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

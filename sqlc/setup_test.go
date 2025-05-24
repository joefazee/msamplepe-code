package db

import (
	"database/sql"
	"os"
	"testing"

	"github.com/timchuks/monieverse/testutils"

	_ "github.com/lib/pq"
	"github.com/timchuks/monieverse/internal/config"
)

var testQueries *Queries
var testDB *sql.DB
var cfg config.Config

const (
	packageName = "db"
)

func TestMain(m *testing.M) {

	setup := testutils.SetupTest(packageName, "../../..", ".env.test")

	testDB = setup.TestDB
	cfg = setup.Config
	testQueries = New(testDB)

	code := m.Run()

	testutils.TeardownTest(setup, packageName)

	os.Exit(code)

}

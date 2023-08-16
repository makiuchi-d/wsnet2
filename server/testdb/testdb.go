package testdb

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	"github.com/dolthub/go-mysql-server/driver"
	"github.com/dolthub/go-mysql-server/memory"
	sqle "github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/jmoiron/sqlx"
)

func init() {
	sql.Register("sqle", driver.New(factory{}, nil))
}

type factory struct{}

func (f factory) Resolve(dbName string, options *driver.Options) (string, sqle.DatabaseProvider, error) {
	memdb := memory.NewDatabase(dbName)
	memdb.EnablePrimaryKeyIndexes()
	provider := memory.NewDBProvider(
		memdb,
		information_schema.NewInformationSchemaDatabase(),
	)
	return dbName, provider, nil
}

func New(dbName string) *sqlx.DB {
	db, err := sql.Open("sqle", dbName)
	if err != nil {
		panic(err)
	}
	dbx := sqlx.NewDb(db, "mysql")
	dbx.MustExec(fmt.Sprintf("USE %s", dbName))
	return dbx
}

func ExtractCreateSql(table, file string) string {
	s, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(fmt.Sprintf("(?im)^CREATE TABLE `?%s[` ][^;]*;", table))
	return re.FindString(string(s))
}

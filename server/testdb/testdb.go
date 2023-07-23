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

func (f factory) Resolve(name string, options *driver.Options) (string, sqle.DatabaseProvider, error) {
	provider := memory.NewDBProvider(
		memory.NewDatabase(name),
		information_schema.NewInformationSchemaDatabase(),
	)
	return name, provider, nil
}

func New(dbName string) *sqlx.DB {
	db, err := sqlx.Open("sqle", dbName)
	if err != nil {
		panic(err)
	}
	db.MustExec(fmt.Sprintf("USE %s", dbName))
	return db
}

func GetCreateSql(table, file string) string {
	s, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(fmt.Sprintf("(?im)^CREATE TABLE `?%s[` ][^;]*;", table))
	return re.FindString(string(s))
}

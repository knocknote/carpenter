package designer

import (
	"database/sql"

	"github.com/knocknote/carpenter/dialect/mysql"
)

func Export(db *sql.DB, schema string, tableNames ...string) (mysql.Tables, error) {
	return mysql.GetTables(db, schema, tableNames...)
}

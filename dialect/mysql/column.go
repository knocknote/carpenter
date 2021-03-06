package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type (
	Column struct {
		TableCatalog           string
		TableSchema            string
		TableName              string
		ColumnName             string
		OrdinalPosition        int32
		ColumnDefault          JsonNullString
		Nullable               string
		DataType               string
		CharacterMaximumLength JsonNullInt64
		CharacterOctetLength   JsonNullInt64
		NumericPrecision       JsonNullInt64
		NumericScale           JsonNullInt64
		CharacterSetName       JsonNullString
		CollationName          JsonNullString
		ColumnType             string
		ColumnKey              string
		Extra                  string
		Privileges             string
		ColumnComment          string
	}
	Columns []*Column
)

func (m *Column) IsPrimary() bool {
	return m.ColumnKey == "PRI"
}

func (m *Column) IsUnique() bool {
	return m.ColumnKey == "UNI"
}

func (m *Column) IsMul() bool {
	return m.ColumnKey == "MUL"
}

func (m *Column) IsAutoIncrement() bool {
	return m.Extra == "auto_increment"
}

func (m *Column) IsNullable() bool {
	return m.Nullable == "YES"
}

func (m *Column) HasDefault() bool {
	return m.ColumnDefault.Valid
}

func (m *Column) HasCharacterSetName() bool {
	return m.CharacterSetName.Valid
}

func (m *Column) HasComment() bool {
	return m.ColumnComment != ""
}

func (m *Column) FormatDefault() string {
	var def string
	switch m.DataType {
	case "char", "varchar", "tinyblob", "blob", "mediumblob", "longblob", "tinytext", "text", "mediumtext", "longtext", "date", "datetime":
		def = QuoteString(m.ColumnDefault.String)
	default:
		def = m.ColumnDefault.String
	}
	return def
}

func (m *Column) CompareCharacterSet(col *Column) bool {
	return reflect.DeepEqual(m.CharacterSetName, col.CharacterSetName)
}

func (m *Column) ToSQL() string {
	token := []string{Quote(m.ColumnName), m.ColumnType}
	if !m.IsNullable() {
		token = append(token, "not null")
	}
	if m.HasDefault() {
		token = append(token, "default", m.FormatDefault())
	}
	if m.IsAutoIncrement() {
		token = append(token, "auto_increment")
	}
	if m.HasComment() {
		token = append(token, "comment", QuoteString(m.ColumnComment))
	}
	return strings.Join(token, " ")
}

func (m *Column) ToAddSQL(pos string) string {
	return fmt.Sprintf("add %s %s", m.ToSQL(), pos)
}

func (m *Column) ToDropSQL() string {
	return fmt.Sprintf("drop %s", Quote(m.ColumnName))
}

func (m *Column) ToModifySQL() string {
	return fmt.Sprintf("modify %s", m.ToSQL())
}

func (m *Column) ToModifyCharsetSQL() string {
	return fmt.Sprintf("modify %s %s %s", Quote(m.ColumnName), m.ColumnType, fmt.Sprintf("character set %s collate %s", m.CharacterSetName.String, m.CollationName.String))
}

func (m Columns) ToSQL() []string {
	sqls := make([]string, 0, len(m))
	for _, col := range m {
		sqls = append(sqls, col.ToSQL())
	}
	return sqls
}

func (m *Column) AppendPos(all Columns) string {
	pos := "first"
	if m.OrdinalPosition > 1 {
		before := m.OrdinalPosition - 1
		for _, v := range all {
			if v.OrdinalPosition != before {
				continue
			}
			pos = fmt.Sprintf("after %s", Quote(v.ColumnName))
			break
		}
	}
	return pos
}

func (m Columns) ToAddSQL(all Columns) []string {
	sqls := make([]string, 0, len(m))
	for _, col := range m {
		sqls = append(sqls, col.ToAddSQL(col.AppendPos(all)))
	}
	return sqls
}

func (m Columns) ToDropSQL() []string {
	sqls := make([]string, 0, len(m))
	for _, col := range m {
		sqls = append(sqls, col.ToDropSQL())
	}
	return sqls
}

func (m Columns) Contains(c *Column) bool {
	for _, v := range m {
		if v.ColumnName == c.ColumnName {
			return true
		}
	}
	return false
}

func (m Columns) GroupByColumnName() map[string]*Column {
	nameMap := make(map[string]*Column, len(m))
	for _, column := range m {
		nameMap[column.ColumnName] = column
	}
	return nameMap
}

func (m Columns) GetSortedColumnNames() []string {
	names := make([]string, 0, len(m))
	for _, column := range m {
		names = append(names, column.ColumnName)
	}
	sort.Strings(names)
	return names
}

func GetColumns(db *sql.DB, schema string) ([]*Column, error) {
	selectCols := []string{
		"TABLE_CATALOG",
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"COLUMN_NAME",
		"ORDINAL_POSITION",
		"COLUMN_DEFAULT",
		"IS_NULLABLE",
		"DATA_TYPE",
		"CHARACTER_MAXIMUM_LENGTH",
		"CHARACTER_OCTET_LENGTH",
		"NUMERIC_PRECISION",
		"NUMERIC_SCALE",
		"CHARACTER_SET_NAME",
		"COLLATION_NAME",
		"COLUMN_TYPE",
		"COLUMN_KEY",
		"EXTRA",
		"PRIVILEGES",
		"COLUMN_COMMENT",
	}
	query := fmt.Sprintf(`select %s from information_schema.columns where TABLE_SCHEMA="%s"`, strings.Join(selectCols, ","), schema)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("err: db.Query failed `%s' for reason %s", query, err)
	}
	defer rows.Close()

	columns := []*Column{}
	for rows.Next() {
		column := &Column{}
		if err := rows.Scan(
			&column.TableCatalog,
			&column.TableSchema,
			&column.TableName,
			&column.ColumnName,
			&column.OrdinalPosition,
			&column.ColumnDefault,
			&column.Nullable,
			&column.DataType,
			&column.CharacterMaximumLength,
			&column.CharacterOctetLength,
			&column.NumericPrecision,
			&column.NumericScale,
			&column.CharacterSetName,
			&column.CollationName,
			&column.ColumnType,
			&column.ColumnKey,
			&column.Extra,
			&column.Privileges,
			&column.ColumnComment,
		); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}

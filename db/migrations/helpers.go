package migrations

import "database/sql"

const DatabaseVersionKeyName = "schema_version"
const DatabaseVersion = 1

// string concatenation is intended in functions below, because tx.Prepare cannot handle this syntax :(
func AddColumn(tx sql.Tx, table string, columnName string, columnType string) error {
	_, err := tx.Exec("ALTER TABLE " + table + " ADD COLUMN " + columnName + " " + columnType)
	return err
}

func RenameColumn(tx sql.Tx, table string, oldColumnName string, newColumnName string, columnType string) error {
	_, err := tx.Exec("ALTER TABLE " + table + " CHANGE COLUMN " + oldColumnName + " " + newColumnName + " " + columnType)
	return err
}

func ModifyColumn(tx sql.Tx, table string, columnName string, columnType string) error {
	_, err := tx.Exec("ALTER TABLE " + table + " MODIFY COLUMN " + columnName + " " + columnType)
	return err
}

func ChangePrimaryKey(tx sql.Tx, table string, primaryKeyString string) error {
	_, err := tx.Exec("ALTER TABLE " + table + " DROP PRIMARY KEY, ADD PRIMARY KEY " + primaryKeyString)
	return err
}

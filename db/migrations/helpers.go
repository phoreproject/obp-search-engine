package migrations

import "database/sql"

const SchemaVersion = "@SCHEMA_VERSION"


func AddColumn(tx sql.Tx, table string, columnName string, columnType string) error {
	statement, err:= tx.Prepare("ALTER TABLE ? ADD ? ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(table, columnName, columnType)
	return err
}


func RenameColumn(tx sql.Tx, table string, oldColumnName string, newColumnName string, columnType string) error {
	statement, err:= tx.Prepare("ALTER TABLE ? CHANGE ? ? ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(table, oldColumnName, newColumnName, columnType)
	return err
}

func ChangeColumnDataType(tx sql.Tx, table string, columnName string, columnType string) error {
	statement, err:= tx.Prepare("ALTER TABLE ? ALTER COLUMN ? ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(table, columnName, columnType)
	return err
}

func ChangePrimaryKey(tx sql.Tx, table string, primaryKeyString string) error {
	_, err := tx.Exec("ALTER TABLE " + table + " DROP PRIMARY KEY, ADD PRIMARY KEY " + primaryKeyString)
	return err
}
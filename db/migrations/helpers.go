package migrations

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
)

const DatabaseVersionKeyName = "schema_version"
const DatabaseVersion = 2 // added 'format' field into items

// string concatenation is intended in functions below, because tx.Prepare cannot handle this syntax :(
func AddColumn(tx sql.Tx, table string, columnName string, columnType string) error {
	str := "ALTER TABLE " + table + " ADD COLUMN " + columnName + " " + columnType
	log.Debugf("AddColumn: %s", str)
	_, err := tx.Exec(str)
	return err
}

func RenameColumn(tx sql.Tx, table string, oldColumnName string, newColumnName string, columnType string) error {
	str := "ALTER TABLE " + table + " CHANGE COLUMN " + oldColumnName + " " + newColumnName + " " + columnType
	log.Debugf("RenameColumn: %s", str)
	_, err := tx.Exec(str)
	return err
}

func ModifyColumn(tx sql.Tx, table string, columnName string, columnType string) error {
	str := "ALTER TABLE " + table + " MODIFY COLUMN " + columnName + " " + columnType
	log.Debugf("ModifyColumn: %s", str)
	_, err := tx.Exec(str)
	return err
}

func DeleteColumn(tx sql.Tx, table string, columnName string) error {
	str := "ALTER TABLE " + table + " DROP COLUMN " + columnName
	log.Debugf("DeleteColumn: %s", str)
	_, err := tx.Exec(str)
	return err
}


func ChangePrimaryKey(tx sql.Tx, table string, primaryKeyString string) error {
	str := "ALTER TABLE " + table + " DROP PRIMARY KEY, ADD PRIMARY KEY " + primaryKeyString
	log.Debugf("ChangePrimaryKey: %s", str)
	_, err := tx.Exec(str)
	return err
}

func RenameTable(tx sql.Tx, oldTableName string, newTableName string) error {
	str := "ALTER TABLE " + oldTableName + " RENAME " + newTableName
	log.Debugf("RenameTable: %s", str)
	_, err := tx.Exec(str)
	return err
}

func UpdateDatabaseVersion(tx sql.Tx, dbVersion int) error {
	log.Debugf("Updating configuration (database version) to %d", dbVersion)
	stmt, err := tx.Prepare("INSERT INTO configuration (uniqueKey, value) VALUES(?, ?) ON DUPLICATE KEY UPDATE value=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(DatabaseVersionKeyName, dbVersion, dbVersion)
	if err != nil {
		log.Debugf("Database update (%d) failed.", dbVersion)
		return err
	}
	log.Debugf("Updated database version to: %d", dbVersion)
	return err
}
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/phoreproject/obp-search-engine/db/migrations"
	"log"
)

type Migration interface {
	Up(db *sql.DB) error
}

var Migrations = []Migration{
	migrations.Migration000{},
}

func Migrate(db *sql.DB) error {
	var configurationKey sql.NullString
	var configurationValue sql.NullInt64
	var schemaVersionInt int
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS configuration (uniqueKey VARCHAR(32) PRIMARY KEY, value TEXT)")
	if err != nil {
		return err
	}

	selectStatement, err := db.Prepare("SELECT * FROM configuration WHERE uniqueKey = ?")
	if err != nil {
		return nil
	}
	defer selectStatement.Close()

	err = selectStatement.QueryRow(migrations.DatabaseVersionKeyName).Scan(&configurationKey, &configurationValue)
	switch {
	case err == sql.ErrNoRows:
		schemaVersionInt = 0
		log.Println("Schema version is missing then expected version is 0")
	case err != nil:
		log.Println(err)
		return err
	default:
		if configurationKey.Valid && configurationValue.Valid &&
			configurationKey.String == migrations.DatabaseVersionKeyName{
			log.Printf("Found schema version %d\n", configurationValue.Int64)
			schemaVersionInt := int(configurationValue.Int64)
			if int64(schemaVersionInt) != configurationValue.Int64 {
				errMsg := fmt.Sprintf("Var %d overflows maximum int type", configurationValue.Int64)
				log.Println(errMsg)
				return errors.New(errMsg)
			}
		} else {
			log.Printf("Schema version is NULL, expected version is 0")
			return err
		}
	}

	for i := schemaVersionInt; i < len(Migrations); i++ {
		err := Migrations[i].Up(db)
		if err != nil {
			log.Printf("Cannot migrate db, because of error: %s", err)
			return err
		}
	}

	return nil
}

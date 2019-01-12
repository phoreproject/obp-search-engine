package db

import (
	"database/sql"
	"github.com/phoreproject/obp-search-engine/db/migrations"
	"log"
	"strconv"
)

type Migration interface {
	Up(db *sql.DB) error
}

var Migrations = []Migration{
	migrations.Migration000{},
}

func Migrate(db *sql.DB) error {
	var schemaVersion sql.NullString
	var schemaVersionInt int
	err := db.QueryRow("SELECT " + migrations.SchemaVersion).Scan(&schemaVersion)
	switch {
	case err != nil:
		log.Println(err)
		return err
	default:
		if schemaVersion.Valid {
			log.Printf("Found schema version %s\n", schemaVersion.String)
			schemaVersionInt, err = strconv.Atoi(schemaVersion.String)
			if err != nil {
				log.Printf("Cannot parse %s into int\n", schemaVersion.String)
				return err
			}
		} else {
			log.Printf("Schema version is NULL, expected version is 0")
			schemaVersionInt = 0
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

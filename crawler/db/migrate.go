package db

import (
	"database/sql"
	"fmt"
	"github.com/phoreproject/obp-search-engine/crawler/db/migrations"
	log "github.com/sirupsen/logrus"
)

type Migration interface {
	Up(db *sql.DB, dbVersion int) error
}

var Migrations = []Migration{
	migrations.Migration000{},
	migrations.Migration001{},
	migrations.Migration002{},
	migrations.Migration003{},
	migrations.Migration004{},
	migrations.Migration005{},
	migrations.Migration006{},
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
		log.Info("Schema version is missing, then expected version is 0")
	case err != nil:
		return err
	default:
		if configurationKey.Valid && configurationValue.Valid &&
			configurationKey.String == migrations.DatabaseVersionKeyName{
			log.Infof("Found schema version %d\n", configurationValue.Int64)
			schemaVersionInt = int(configurationValue.Int64)
			if int64(schemaVersionInt) != configurationValue.Int64 {
				return fmt.Errorf("var %d overflows maximum int type", configurationValue.Int64)
			}
		} else {
			log.Error("Schema version is NULL, expected version is 0")
			return err
		}
	}

	if schemaVersionInt == len(Migrations) {
		log.Info("Nothing to do. Database is updated.")
		return nil
	} else if schemaVersionInt > len(Migrations) {
		return fmt.Errorf("current schema version is higher (%d) than code version (%d)", schemaVersionInt, len(Migrations))
	}


	for i := schemaVersionInt; i < len(Migrations); i++ {
		log.Debugf("Migration%d starting", i)
		err := Migrations[i].Up(db, i+1)
		if err != nil {
			return err
		}
	}
	log.Infof("Migrated to version %d", len(Migrations))

	return nil
}

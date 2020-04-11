package migrations

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type Migration004 struct{}

func (Migration004) Up(db *sql.DB, dbVersion int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	const itemsTableName = "items"
	if err = AddColumn(*tx, itemsTableName, "blocked", "TINYINT(1) AFTER normalizedPrice"); err != nil {
		return err
	}

	return UpdateDatabaseVersion(*tx, dbVersion)
}

package migrations

import (
	"context"
	"database/sql"
)

type Migration001 struct{}

func (Migration001) Up(db *sql.DB, dbVersion int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	// old table name
	const oldTableName = "moderatorIdsPerItem"
	const newTableName = "moderatorIdsPerNode"

	if err = RenameTable(*tx, oldTableName, newTableName); err != nil {
		return err
	}

	if err = ChangePrimaryKey(*tx, newTableName, "(peerID, moderatorID)"); err != nil {
		return err
	}

	if err = DeleteColumn(*tx, newTableName, "itemDataBaseID"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO configuration (uniqueKey, value) VALUES(?, ?) ON DUPLICATE KEY UPDATE value=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(DatabaseVersionKeyName, dbVersion, dbVersion)
	if err != nil {
		return err
	}

	return err
}

package migrations

import (
	"context"
	"database/sql"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Migration000 struct{}

func (Migration000) Up(db *sql.DB, dbVersion int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Panic(p)
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	// add new column into nodes
	log.Debugf("Migrating nodes table")
	const nodeTableName = "nodes"
	if err = AddColumn(*tx, nodeTableName, "userAgent", "VARCHAR(50) FIRST"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "verifiedModerator", "TINYINT(1) AFTER moderator"); err != nil {
		return err
	}

	// rename banned to blocked
	if err = RenameColumn(*tx, nodeTableName, "banned", "blocked", "TINYINT(1) DEFAULT 0"); err != nil {
		return err
	}

	// add avatar hashes columns
	if err = AddColumn(*tx, nodeTableName, "avatarTinyHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "avatarSmallHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "avatarMediumHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "avatarOriginalHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "avatarLargeHash", "VARCHAR(50)"); err != nil {
		return err
	}

	//add header hashes columns
	if err = AddColumn(*tx, nodeTableName, "headerTinyHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "headerSmallHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "headerMediumHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "headerOriginalHash", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "headerLargeHash", "VARCHAR(50)"); err != nil {
		return err
	}

	log.Debugf("Migrating items table")
	// add new columns into items
	const itemsTableName = "items"
	if err = AddColumn(*tx, itemsTableName, "id", "INT FIRST"); err != nil {
		return err
	}

	var configurationValue sql.NullInt64
	selectCnt := "SELECT COUNT(*) from " + itemsTableName
	log.Debugf("Counting items: %s", selectCnt)
	err = tx.QueryRow(selectCnt).Scan(&configurationValue)
	if err != nil {
		return err
	}
	if !configurationValue.Valid {
		return errors.New("result of SELECT COUNT(*) from items isn't valid")
	}

	for i := 0; i < int(configurationValue.Int64); i++ {
		err = func() error {
			updatingItemId := "UPDATE " + itemsTableName + " SET id = ? WHERE id IS NULL LIMIT 1"
			log.Debugf("Updating item id: %s", updatingItemId)
			stmt, err := tx.Prepare(updatingItemId)
			if err != nil {
				return err
			}
			defer stmt.Close()

			_, err = tx.Stmt(stmt).Exec(i)
			return err
		}()
		if err != nil {
			return err
		}
	}

	if err = ChangePrimaryKey(*tx, itemsTableName, "(id)"); err != nil {
		return err
	}
	autoIncStart := strconv.FormatInt(configurationValue.Int64, 10)
	if err = ModifyColumn(*tx, itemsTableName, "id", "INT NOT NULL AUTO_INCREMENT, AUTO_INCREMENT="+autoIncStart); err != nil {
		return err
	}

	if err = RenameColumn(*tx, itemsTableName, "owner", "peerID", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "score", "TINYINT AFTER peerID"); err != nil {
		return err
	}
	if err = ModifyColumn(*tx, itemsTableName, "thumbnail", "VARCHAR(260)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "priceModifier", "INT AFTER priceCurrency"); err != nil {
		return err
	}
	if err = RenameColumn(*tx, itemsTableName, "rating", "averageRating", "DECIMAL(3,2)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "ratingCount", "INT AFTER averageRating"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "coinType", "VARCHAR(20) AFTER ratingCount"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "coinDivisibility", "INT AFTER coinType"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "normalizedPrice", "DECIMAL(40, 20) AFTER coinDivisibility"); err != nil {
		return err
	}
	if err = ModifyColumn(*tx, itemsTableName, "categories", "VARCHAR(410) AFTER tags"); err != nil {
		return err
	}
	if err = ModifyColumn(*tx, itemsTableName, "contractType", "VARCHAR(20) AFTER categories"); err != nil {
		return err
	}
	if err = ModifyColumn(*tx, itemsTableName, "description", "TEXT AFTER contractType"); err != nil {
		return err
	}

	// add new tables
	log.Debugf("Creating table moderators")
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS moderators (id VARCHAR(50) NOT NULL, type VARCHAR(16), " +
		"isVerified TINYINT(1) DEFAULT 0, PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	log.Debugf("Creating table moderatorIdsPerItem")
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS moderatorIdsPerItem (peerID VARCHAR(50) NOT NULL, " +
		"itemDataBaseID INT NOT NULL, moderatorID VARCHAR(50) NOT NULL, PRIMARY KEY(peerID, itemDataBaseID, moderatorID))")
	if err != nil {
		return err
	}

	log.Debugf("Updating configuration (database version) to %d", dbVersion)
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

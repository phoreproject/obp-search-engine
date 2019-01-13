package migrations

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

type Migration000 struct{}

func (Migration000) Up(db *sql.DB) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// add new column into nodes
	const nodeTableName = "nodes"
	if err = AddColumn(*tx, nodeTableName, "userAgent", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, nodeTableName, "verifiedModerator", "TINYINT(1) DEFAULT 0"); err != nil {
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
		return errselec
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

	// add new columns into items
	const itemsTableName = "items"
	if err = AddColumn(*tx, itemsTableName, "id", "INT"); err != nil {
		return err
	}

	var configurationValue sql.NullInt64
	err = tx.QueryRow("SELECT COUNT(*) from " + itemsTableName).Scan(&configurationValue)
	if err != nil {
		return err
	}
	if !configurationValue.Valid {
		return errors.New("result of SELECT COUNT(*) from items isn't valid")
	}

	for i := 0; i < int(configurationValue.Int64); i++ {
		err = func() error {
			stmt, err := tx.Prepare("UPDATE " + itemsTableName + " SET id = ? WHERE id IS NULL LIMIT 1")
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
	if err = ModifyColumn(*tx, itemsTableName, "id", "INT NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=" + autoIncStart); err != nil {
		return err
	}

	if err = RenameColumn(*tx, itemsTableName, "owner", "PeerID", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "score", "TINYINT"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "peerID", "VARCHAR(50)"); err != nil {
		return err
	}
	if err = ModifyColumn(*tx, itemsTableName, "thumbnail", "VARCHAR(260)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "priceModifier", "INT"); err != nil {
		return err
	}
	if err = RenameColumn(*tx, itemsTableName, "rating", "averageRating", "DECIMAL(3,2)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "ratingCount", "INT"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "coinType", "VARCHAR(20)"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "coinDivisibility", "INT"); err != nil {
		return err
	}
	if err = AddColumn(*tx, itemsTableName, "normalizedPrice", "DECIMAL(40, 20)"); err != nil {
		return err
	}

	// add new tables
	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS moderators (id VARCHAR(50) NOT NULL, type VARCHAR(16), " +
		"isVerified TINYINT(1) DEFAULT 0, PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS moderatorIdsPerItem (peerID VARCHAR(50) NOT NULL, " +
		"itemDataBaseID INT NOT NULL, moderatorID VARCHAR(50) NOT NULL, PRIMARY KEY(peerID, itemDataBaseID, moderatorID))")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE configuration SET value = ? WHERE uniqueKey = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(1, DatabaseVersionKeyName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

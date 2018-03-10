package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/phoreproject/obp-search-engine/crawling"
)

// SQLDatastore represents a datastore for the crawler implemented using Redis
type SQLDatastore struct {
	db *sql.DB
}

// NewSQLDatastore creates a new datastore given MySQL connection info
func NewSQLDatastore(db *sql.DB) (*SQLDatastore, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS nodes (id VARCHAR(50) NOT NULL, lastUpdated DATETIME, PRIMARY KEY (id))")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS items (owner VARCHAR(50), hash VARCHAR(50) NOT NULL, slug VARCHAR(70), title VARCHAR(140), tags VARCHAR(410), description VARCHAR(50000), thumbnail VARCHAR(160), language VARCHAR(20), priceAmount BIGINT, priceCurrency VARCHAR(10), categories VARCHAR(410), nsfw TINYINT(1), contractType VARCHAR(20), rating DECIMAL(3, 2), PRIMARY KEY (hash))")
	if err != nil {
		return nil, err
	}
	return &SQLDatastore{db: db}, nil
}

// GetNextNode gets the next node from the database
func (d *SQLDatastore) GetNextNode() (*crawling.Node, error) {
	r := d.db.QueryRow("SELECT id, lastUpdated FROM nodes ORDER BY lastUpdated ASC LIMIT 1")
	node := crawling.Node{}
	err := r.Scan(&node.ID, &node.LastCrawled)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// SaveNode saves a node to the database
func (d *SQLDatastore) SaveNode(n crawling.Node) error {
	insertStatement, err := d.db.Prepare("INSERT INTO nodes (id, lastUpdated) VALUES (?, NOW()) ON DUPLICATE KEY UPDATE lastUpdated=NOW()")
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(insertStatement).Exec(n.ID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

// AddUninitializedNodes adds nodes to the queue to be crawled
func (d *SQLDatastore) AddUninitializedNodes(nodes []crawling.Node) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	for n := range nodes {
		fmt.Printf("Added %s\n", nodes[n].ID)
		insertStatement, err := d.db.Prepare("INSERT IGNORE INTO nodes (id, lastUpdated) VALUES (?, '2000-01-01 00:00:00')")
		if err != nil {
			return err
		}

		_, err = tx.Stmt(insertStatement).Exec(nodes[n].ID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

// GetNode gets a node's information from the datastore
func (d *SQLDatastore) GetNode(nodeID string) (*crawling.Node, error) {
	s, err := d.db.Prepare("SELECT id, lastUpdated FROM nodes WHERE id=?")
	if err != nil {
		return nil, err
	}
	r := s.QueryRow(nodeID)
	node := &crawling.Node{}
	err = r.Scan(&node.ID, &node.LastCrawled)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// AddItemsForNode updates a node with the following items
func (d *SQLDatastore) AddItemsForNode(owner string, items []crawling.Item) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	s, err := tx.Prepare("DELETE FROM items WHERE owner = ?")
	if err != nil {
		return err
	}

	_, err = s.Exec(owner)
	if err != nil {
		return err
	}

	for i := range items {
		s, err = tx.Prepare("INSERT INTO items (owner, hash, slug, title, tags, description, thumbnail, language, priceAmount, priceCurrency, categories, nsfw, contractType, rating) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return err
		}

		_, err = s.Exec(
			owner,
			items[i].Hash,
			items[i].Slug,
			items[i].Title,
			"",
			items[i].Description,
			items[i].Thumbnail.Tiny+","+items[i].Thumbnail.Small+","+items[i].Thumbnail.Medium,
			items[i].Language,
			items[i].Price.Amount,
			items[i].Price.CurrencyCode,
			strings.Join(items[i].Categories, ","),
			items[i].NSFW,
			items[i].ContractType,
			items[i].AverageRating,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

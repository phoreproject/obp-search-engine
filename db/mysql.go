package db

import (
	"database/sql"
	"fmt"

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

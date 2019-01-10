package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/phoreproject/obp-search-engine/crawling"
)

// start mysql container
// docker run -p 3306:3306 --name mysqlTest -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_ROOT_HOST=% -e MYSQL_DATABASE=obpsearch -d mysql/mysql-server

// SQLDatastore represents a datastore for the crawler implemented using Redis
type SQLDatastore struct {
	db *sql.DB
}

// NewSQLDatastore creates a new datastore given MySQL connection info
func NewSQLDatastore(db *sql.DB) (*SQLDatastore, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS nodes (userAgent VARCHAR(50), id VARCHAR(50) NOT NULL, lastUpdated DATETIME, " +
		"name VARCHAR(40), handle VARCHAR(40), location VARCHAR(40), nsfw TINYINT(1), vendor TINYINT(1), moderator TINYINT(1), " +
		"verifiedModerator TINYINT(1) DEFAULT 0, about VARCHAR(10000), shortDescription VARCHAR(160), followerCount INT, " +
		"followingCount INT, listingCount INT, postCount INT, ratingCount INT, averageRating DECIMAL(3, 2), listed TINYINT(1) DEFAULT 0, " +
		"blocked TINYINT(1) DEFAULT 0, PRIMARY KEY (id))")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS items (owner VARCHAR(50), score TINYINT, hash VARCHAR(50) NOT NULL," +
		" slug VARCHAR(70), title VARCHAR(140), tags VARCHAR(410), categories VARCHAR(410), contractType VARCHAR(20)," +
		" description TEXT, thumbnail VARCHAR(260), language VARCHAR(20), priceAmount BIGINT, priceCurrency VARCHAR(10)," +
		" priceModifier INT, nsfw TINYINT(1), averageRating INT, ratingCount INT, coinType VARCHAR(20), coinDivisibility INT," +
		" normalizedPrice DECIMAL(40, 20), PRIMARY KEY (hash))")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS moderators (id VARCHAR(50) NOT NULL, type VARCHAR(16), " +
		"isVerified TINYINT(1) DEFAULT 0), PRIMARY KEY(id)")
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

// SaveNodeUninitialized saves a node to the database without extra data
func (d *SQLDatastore) SaveNodeUninitialized(n crawling.Node) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	insertStatement, err := tx.Prepare("INSERT INTO nodes (id, lastUpdated, userAgent) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE lastUpdated=NOW(), userAgent=?")
	if err != nil {
		return err
	}
	defer insertStatement.Close()

	_, err = tx.Stmt(insertStatement).Exec(n.ID, n.UserAgent)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

// SaveNode saves a node to the database
func (d *SQLDatastore) SaveNode(n crawling.Node) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	insertStatement, err := tx.Prepare("INSERT INTO nodes (id, lastUpdated, userAgent, name, handle, location, nsfw, vendor, " +
		"moderator, about, shortDescription, followerCount, followingCount, listingCount, postCount, ratingCount, averageRating) " +
		"VALUES (?, NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE lastUpdated=NOW(), userAgent=?, name=?, " +
		"handle=?, location=?, nsfw=?, vendor=?, moderator=?, about=?, shortDescription=?, followerCount=?, followingCount=?, " +
		"listingCount=?, postCount=?, ratingCount=?, averageRating=?")
	if err != nil {
		return err
	}
	defer insertStatement.Close()

	_, err = tx.Stmt(insertStatement).Exec(
		n.ID,
		n.UserAgent,
		n.Profile.Name,
		n.Profile.Handle,
		n.Profile.Location,
		n.Profile.Nsfw,
		n.Profile.Vendor,
		n.Profile.Moderator,
		n.Profile.About,
		n.Profile.ShortDescription,
		n.Profile.Stats.FollowerCount,
		n.Profile.Stats.FollowingCount,
		n.Profile.Stats.ListingCount,
		n.Profile.Stats.PostCount,
		n.Profile.Stats.RatingCount,
		n.Profile.Stats.AverageRating,

		// on duplicated
		n.UserAgent,
		n.Profile.Name,
		n.Profile.Handle,
		n.Profile.Location,
		n.Profile.Nsfw,
		n.Profile.Vendor,
		n.Profile.Moderator,
		n.Profile.About,
		n.Profile.ShortDescription,
		n.Profile.Stats.FollowerCount,
		n.Profile.Stats.FollowingCount,
		n.Profile.Stats.ListingCount,
		n.Profile.Stats.PostCount,
		n.Profile.Stats.RatingCount,
		n.Profile.Stats.AverageRating,
	)
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
		err = func() error {
			fmt.Printf("Added %s\n", nodes[n].ID)
			insertStatement, err := d.db.Prepare("INSERT IGNORE INTO nodes (id, lastUpdated) VALUES (?, '2000-01-01 00:00:00')")
			if err != nil {
				return err
			}
			defer insertStatement.Close()

			_, err = tx.Stmt(insertStatement).Exec(nodes[n].ID)
			return err
		}()
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
	defer s.Close()
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
	defer s.Close()

	_, err = s.Exec(owner)
	if err != nil {
		return err
	}

	for i := range items {
		err = func() error {
			s, err = tx.Prepare("INSERT INTO items (owner, hash, score, slug, title, tags, categories, contractType, " +
				"description, thumbnail, language, priceAmount, priceCurrency, priceModifier, nsfw, averageRating, ratingCount, " +
				"coinType, coinDivisibility, normalizedPrice) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE " +
				"score=?, slug=?, title=?, tags=?, categories=?, contractType=?, description=?, thumbnail=?, language=?, " +
				"priceAmount=?, priceCurrency=?, priceModifier=?, nsfw=?, averageRating=?, ratingCount=?, coinType=?," +
				"coinDivisibility=?, normalizedPrice=?")
			if err != nil {
				return err
			}

			defer s.Close()

			_, err = s.Exec(
				owner,
				items[i].Hash,

				items[i].Score,
				items[i].Slug,
				items[i].Title,
				strings.Join(items[i].Tags, ","),
				strings.Join(items[i].Categories, ","),
				items[i].ContractType,
				items[i].Description,
				items[i].Thumbnail.Tiny+","+items[i].Thumbnail.Small+","+items[i].Thumbnail.Medium+","+items[i].Thumbnail.Original+","+items[i].Thumbnail.Large,
				items[i].Language,
				items[i].Price.Amount,
				items[i].Price.CurrencyCode,
				items[i].Price.Modifier,
				items[i].NSFW,
				items[i].AverageRating,
				items[i].RatingCount,
				items[i].CoinType,
				items[i].CoinDivisibility,
				items[i].NormalizedPrice,

				// on duplicate repeat
				items[i].Score,
				items[i].Slug,
				items[i].Title,
				strings.Join(items[i].Tags, ","),
				strings.Join(items[i].Categories, ","),
				items[i].ContractType,
				items[i].Description,
				items[i].Thumbnail.Tiny+","+items[i].Thumbnail.Small+","+items[i].Thumbnail.Medium+","+items[i].Thumbnail.Original+","+items[i].Thumbnail.Large,
				items[i].Language,
				items[i].Price.Amount,
				items[i].Price.CurrencyCode,
				items[i].Price.Modifier,
				items[i].NSFW,
				items[i].AverageRating,
				items[i].RatingCount,
				items[i].CoinType,
				items[i].CoinDivisibility,
				items[i].NormalizedPrice,
			)
			return err
		}()

		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

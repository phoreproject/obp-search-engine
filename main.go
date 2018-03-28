package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
)

func main() {

	databaseURL := flag.String("mysql", "root@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	flag.Parse()

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		panic(err)
	}

	d, err := db.NewSQLDatastore(database)
	if err != nil {
		panic(err)
	}

	r := rpc.NewRPC(*rpcURL)

	c := &crawling.Crawler{RPCInterface: r, DB: d}

	config, err := r.GetConfig()
	if err != nil {
		log.Fatal("You need to run openbazaard. Please check: https://github.com/phoreproject/openbazaar-go")
	}

	profile, err := r.GetProfile(config.PeerID)
	if err != nil {
		panic(err)
	}

	// add ourselves
	err = c.DB.SaveNodeUninitialized(crawling.Node{ID: config.PeerID, Connections: []string{}, Profile: profile})
	if err != nil {
		panic(err)
	}

	for {
		done := make(chan bool)
		go func() {
			nodeID, err := c.CrawlOnce()

			if err != nil {
				done <- true
				return
			}

			fmt.Printf("Crawling %s\n", nodeID)

			_, err = c.RPCInterface.GetUserAgent(nodeID)
			if err != nil {
				done <- true
				fmt.Printf("Could not access node %s. ignoring\n", nodeID)
				return
			}

			items, err := c.RPCInterface.GetItems(nodeID)
			if err != nil {
				done <- true
				return
			}

			fmt.Printf("Found %d items.\n", len(items))

			err = c.DB.AddItemsForNode(nodeID, items)
			if err != nil {
				done <- true
				return
			}
			done <- true
		}()
		select {
		case <-done:
			break
		case <-time.After(10 * time.Second):
			break
		}
	}
}

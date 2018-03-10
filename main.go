package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
)

func main() {

	databaseURL := flag.String("mysql", "root@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	flag.Parse()

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true")
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
		panic(err)
	}

	// add ourselves
	err = c.DB.SaveNode(crawling.Node{ID: config.PeerID, Connections: []string{}})
	if err != nil {
		panic(err)
	}

	for {
		done := make(chan bool)
		timeout := make(chan bool)
		go func() {
			nodeID, err := c.CrawlOnce()

			if err != nil {
				panic(err)
			}

			if len(nodeID) < 40 {
				nodeID = config.PeerID
			}

			fmt.Printf("Crawling %s\n", nodeID)

			items, err := c.RPCInterface.GetItems(nodeID)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Found %d items.\n", len(items))

			err = c.DB.AddItemsForNode(nodeID, items)
			if err != nil {
				panic(err)
			}
			done <- true
		}()
		go func() {
			time.Sleep(time.Second * 10)
			timeout <- true
		}()
		select {
		case <-done:
			break
		case <-time.After(10 * time.Second):
			break
		}
	}
}

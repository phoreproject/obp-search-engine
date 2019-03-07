package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
	"log"
	"sync"
	"time"
)

func main() {
	// url format is user:password@protocol(address:port)/db_name
	databaseURL := flag.String("mysql", "root@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	skipMigration := flag.Bool("skipMigration", false, "skip database migration to the newest version on start")
	flag.Parse()

	CHUNK_SIZE := 100
	MAX_PARALLEL_COROUTINE := 10

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		panic(err)
	}

	d, err := db.NewSQLDatastore(database, !(*skipMigration))
	if err != nil {
		panic(err)
	}

	r := rpc.NewRPC(*rpcURL)

	c := &crawling.Crawler{RPCInterface: r, DB: d}

	config, err := r.GetConfig()
	if err != nil {
		log.Printf("You need to run openbazaard. Please check: https://github.com/phoreproject/openbazaar-go")
		panic(err)
	}

	profile, err := r.GetProfile(config.PeerID)
	if err != nil {
		panic(err)
	}

	userAgent, err := c.RPCInterface.GetUserAgentFromIPNS(config.PeerID)
	if err != nil {
		panic(err)
	}

	// add ourselves
	err = c.DB.SaveNodeUninitialized(crawling.Node{ID: config.PeerID, UserAgent: userAgent, Connections: []string{}, Profile: profile})
	if err != nil {
		panic(err)
	}
	crawlerRound := 1
	for {
		startTime := time.Now()
		processedCnt := 0
		done := make(chan bool)
		go func() {
			lastNodeID := ""
			for {
				// get next chunk of nodes from database
				nodesIDs, err := c.DB.GetNextNodesChan(lastNodeID, CHUNK_SIZE)
				if err != nil {
					done <- true
					return
				}

				atLeastOneNode := false
				for nodeID := range nodesIDs {
					atLeastOneNode = true

					// create list of at max MAX_PARALLEL_COROUTINE nodes to start parallel coroutines
					var lastNodes []string
					lastNodes = append(lastNodes, nodeID)
					nodesLen := MAX_PARALLEL_COROUTINE - 1
					for nodeID = range nodesIDs {
						lastNodes = append(lastNodes, nodeID)
						nodesLen--
						if nodesLen <= 0 {
							break
						}
					}
					lastNodeID = nodeID

					// start len(lastNodes) parallel coroutines
					var wg sync.WaitGroup
					wg.Add(len(lastNodes))
					processedCnt += len(lastNodes)
					for i := range lastNodes {
						go func(localNodeID string) {
							defer wg.Done()
							fmt.Printf("Processing node with id: %s\n", localNodeID)
							err := c.CrawlNode(localNodeID)
							if err != nil {
								fmt.Println(err)
								return
							}

							fmt.Printf("Crawling items for localNodeID: %s\n", localNodeID)
							items, err := c.RPCInterface.GetItems(localNodeID)
							if err != nil {
								fmt.Println(err)
								return
							}

							err = c.DB.AddItemsForNode(localNodeID, items)
							if err != nil {
								fmt.Println(err)
								return
							}
						}(lastNodes[i])
					}
					wg.Wait()
				}

				if !atLeastOneNode {
					break
				}
			}
			done <- true
		}()

		select {
		case <-done:
			crawlerRound++
			log.Printf("Crawler round %d took %s and processed %d items", crawlerRound, time.Since(startTime), processedCnt)
			processedCnt = 0
		}
	}
}

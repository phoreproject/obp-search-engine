package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	// configure logger output format
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	// configure logger writers
	logFile, err := os.Create("crawler.log")
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, logFile)) // write to file and stderr

	// url format is user:password@protocol(address:port)/db_name
	databaseURL := flag.String("mysql", "root@tcp(127.0.0.1:3306)/obpsearch", "database url used to connect to MySQL database")
	rpcURL := flag.String("rpc", "127.0.0.1:5002", "rpc url used to connect to Phore Marketplace")
	skipMigration := flag.Bool("skipMigration", false, "skip database migration to the newest version on start")
	verbose := flag.Bool("verbose", false, "use more verbose logging")
	chunkSize := flag.Int("chunkSize", 100, "Maximum database select chunk size")
	maxParallelCorutine := flag.Int("maxCoroutine", 10, "Maximum number of parallel connections")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Info("Using verbose logging!")
	}

	log.Debugf("Starting app with chunk size %d, and max parallel corutine cnt %d", chunkSize, maxParallelCorutine)

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		log.Panic(err)
	}

	d, err := db.NewSQLDatastore(database, !(*skipMigration))
	if err != nil {
		log.Panic(err)
	}

	r := rpc.NewRPC(*rpcURL)

	c := &crawling.Crawler{RPCInterface: r, DB: d}

	config, err := r.GetConfig()
	if err != nil {
		log.Panic("You need to run openbazaard. Please check: https://github.com/phoreproject/openbazaar-go")
	}

	profile, err := r.GetProfile(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	userAgent, err := c.RPCInterface.GetUserAgentFromIPNS(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	// add ourselves
	err = c.DB.SaveNodeUninitialized(crawling.Node{ID: config.PeerID, UserAgent: userAgent, Connections: []string{}, Profile: profile})
	if err != nil {
		log.Panic(err)
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
				nodesIDs, err := c.DB.GetNextNodesChan(lastNodeID, *chunkSize)
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
					nodesLen := *maxParallelCorutine - 1
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
					processingStart := time.Now()
					for i := range lastNodes {
						go func(localNodeID string) {
							defer wg.Done()
							log.Debugf("Processing node with id: %s\n", localNodeID)
							err := c.CrawlNode(localNodeID)
							if err != nil {
								log.Error(err)
								return
							}

							log.Debugf("Crawling items for localNodeID: %s\n", localNodeID)
							items, err := c.RPCInterface.GetItems(localNodeID)
							if err != nil {
								log.Error(err)
								return
							}

							err = c.DB.AddItemsForNode(localNodeID, items)
							if err != nil {
								log.Error(err)
								return
							}
						}(lastNodes[i])
					}
					wg.Wait()
					log.Debugf("Processing of %d nodes took %s.", len(lastNodes), time.Since(processingStart))
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
			log.Debugf("Crawler round %d took %s and processed %d items", crawlerRound, time.Since(startTime), processedCnt)
			processedCnt = 0
		}
	}
}

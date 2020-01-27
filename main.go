package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/phoreproject/obp-search-engine/crawling"
	"github.com/phoreproject/obp-search-engine/db"
	"github.com/phoreproject/obp-search-engine/rpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

const (
	BanThreshold   = 0.5
	AllowThreshold = 0.1
)

func checkListings(url string, items []crawling.Item) ([]int, error) {
	c := &http.Client{Timeout: time.Second * 10}

	s, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "http://"+path.Join(url, "checkListings"), bytes.NewBuffer(s))
	req.Header.Set("Content-type", "application/json")
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Server returned %s", resp.Status))
	}

	var response []int
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func crawlerMainLoop(maxParallelCoroutines int, chunkSize int, httpClassifierUrl string, crawler *crawling.Crawler) {
	crawlerRound := 1
	for {
		startTime := time.Now()
		processedCnt := 0
		done := make(chan bool)
		go func() {
			lastNodeID := ""
			for {
				// get next chunk of nodes from database
				nodesIDs, err := crawler.DB.GetNextNodesChan(lastNodeID, chunkSize)
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
					nodesLen := maxParallelCoroutines - 1
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
							err := crawler.CrawlNode(localNodeID)
							if err != nil {
								log.Error(err)
								return
							}

							log.Debugf("Crawling items for localNodeID: %s\n", localNodeID)
							items, err := crawler.RPCInterface.GetItems(localNodeID)
							if err != nil {
								log.Error(err)
								return
							}

							if httpClassifierUrl != "" && len(items) > 0 {
								log.Debugf("Checking listings status using external service")
								output, err := checkListings(httpClassifierUrl, items)
								if err != nil {
									log.Warning(err)
								}
								log.Debugf("Check finished with %s", output)

								if len(output) == len(items) {
									bannedCnt := 0
									for index, blocked := range output {
										if blocked != 0 {
											bannedCnt += 1
										}
										items[index].Blocked = blocked > 0
									}

									if bannedCnt == 0 || float64(len(items))/float64(bannedCnt) < AllowThreshold { // automatically list owner
										// automatically list stores with low level of banned listings
										err := crawler.DB.UpdateNodeStatus(localNodeID, "listed", true)
										if err != nil {
											log.Warning("Cannot update node status listed = true for node %s", localNodeID)
											log.Error(err)
										} else {
											log.Debugf("Updated node status listed = true for node %s", localNodeID)
										}
									} else if float64(len(items))/float64(bannedCnt) > BanThreshold {
										// automatically ban also entire owner store
										err := crawler.DB.UpdateNodeStatus(localNodeID, "blocked", true)
										if err != nil {
											log.Warning("Cannot update node status banned = true for node %s", localNodeID)
											log.Error(err)
										} else {
											log.Debugf("Updated node status banned = true for node %s", localNodeID)
										}
									} else {
										log.Debugf("Node %s banned listing threshold in range (%f < %f < %f) - cannot decide automatically",
											localNodeID, AllowThreshold, float64(len(items))/float64(bannedCnt), BanThreshold)
									}
								}
							}

							err = crawler.DB.AddItemsForNode(localNodeID, items)
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
	httpClassifierUrl := flag.String("httpClassifierUrl", "", "Service url for classifying listings")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Info("Using verbose logging!")
	}

	log.Debugf("Starting app with chunk size %d, and max parallel corutine cnt %d", *chunkSize, *maxParallelCorutine)

	database, err := sql.Open("mysql", *databaseURL+"?parseTime=true&interpolateParams=true")
	if err != nil {
		log.Panic(err)
	}

	d, err := db.NewSQLDatastore(database, !(*skipMigration))
	if err != nil {
		log.Panic(err)
	}

	r := rpc.NewRPC(*rpcURL)

	crawler := &crawling.Crawler{RPCInterface: r, DB: d}

	config, err := r.GetConfig()
	if err != nil {
		log.Panic("You need to run openbazaard. Please check: https://github.com/phoreproject/openbazaar-go")
	}

	profile, err := r.GetProfile(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	userAgent, err := crawler.RPCInterface.GetUserAgentFromIPNS(config.PeerID)
	if err != nil {
		log.Panic(err)
	}

	// add ourselves
	err = crawler.DB.SaveNodeUninitialized(crawling.Node{ID: config.PeerID, UserAgent: userAgent, Connections: []string{}, Profile: profile})
	if err != nil {
		log.Panic(err)
	}

	crawlerMainLoop(*maxParallelCorutine, *chunkSize, *httpClassifierUrl, crawler)
}

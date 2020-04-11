package crawling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	BanThreshold   = 0.5
	AllowThreshold = 0.1
)

func checkListings(url string, items []Item) ([]int, error) {
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

// Crawler handles crawling through each node in OB
type Crawler struct {
	DB                Datastore
	RPCInterface      RPCInterface
	WorkerQueue       chan chan Node
	MaxCoroutineCnt   int
	ChunkSize         int
	HttpClassifierUrl string
}

// CrawlNode runs the crawler for one step
func (c Crawler) CrawlNode(nextNodeID string) error {
	nextNode := Node{ID: nextNodeID, LastCrawled: time.Now()}

	connections, err := c.RPCInterface.GetConnections(nextNode.ID)
	if err != nil {
		return err
	}

	nodes := []Node{}
	for i := range connections {
		if connections[i] != nextNode.ID && len(connections[i]) > 40 && connections[i][0] != ' ' {
			nodes = append(nodes, Node{ID: connections[i], Connections: []string{}, LastCrawled: time.Date(2017, 12, 13, 0, 0, 0, 0, time.Local)})
		}
	}
	nextNode.Connections = connections
	if err := c.DB.AddUninitializedNodes(nodes); err != nil {
		return err
	}

	userAgent, err := c.RPCInterface.GetUserAgentFromIPNS(nextNode.ID)
	if err != nil {
		return err
	} else if strings.Contains(userAgent, nextNode.ID) { // marketplace returns
		return fmt.Errorf("Could not access node %s. Ignoring.\n  IPNS returned: %s", nextNode.ID, userAgent)
	}
	nextNode.UserAgent = userAgent

	profile, err := c.RPCInterface.GetProfile(nextNode.ID)
	if profile != nil && profile.Stats != nil {
		nextNode.Profile = profile

		log.Debugf("Saving node %s...\n", nextNode.ID)
		if err := c.DB.SaveNode(nextNode); err != nil {
			return err
		}
	} else {
		log.Debugf("Saving empty node %s...\n", nextNode.ID)
		if err := c.DB.SaveNodeUninitialized(nextNode); err != nil {
			return err
		}
	}
	return nil
}

func (c Crawler) classify(nodeID string, items []Item) {
	log.Debugf("Checking listings status using external service")
	output, err := checkListings(c.HttpClassifierUrl, items)
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
			err := c.DB.UpdateNodeStatus(nodeID, "listed", true)
			if err != nil {
				log.Warning("Cannot update node status listed = true for node %s", nodeID)
				log.Error(err)
			} else {
				log.Debugf("Updated node status listed = true for node %s", nodeID)
			}
		} else if float64(len(items))/float64(bannedCnt) > BanThreshold {
			// automatically ban also entire owner store
			err := c.DB.UpdateNodeStatus(nodeID, "blocked", true)
			if err != nil {
				log.Warning("Cannot update node status banned = true for node %s", nodeID)
				log.Error(err)
			} else {
				log.Debugf("Updated node status banned = true for node %s", nodeID)
			}
		} else {
			log.Debugf("Node %s banned listing threshold in range (%f < %f < %f) - cannot decide automatically",
				nodeID, AllowThreshold, float64(len(items))/float64(bannedCnt), BanThreshold)
		}
	}
}

func (c Crawler) ProcessOneNodeSync(nodeID string) {
	log.Debugf("Processing node with id: %s\n", nodeID)
	err := c.CrawlNode(nodeID)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Crawling items for nodeID: %s\n", nodeID)
	items, err := c.RPCInterface.GetItems(nodeID)
	if err != nil {
		log.Error(err)
		return
	}

	if c.HttpClassifierUrl != "" && len(items) > 0 {
		c.classify(nodeID, items)
	}

	err = c.DB.AddItemsForNode(nodeID, items)
	if err != nil {
		log.Error(err)
		return
	}
}

func (c Crawler) ProcessOneNodeAsync(nodeID string, wg *sync.WaitGroup) {
	defer wg.Done()
	c.ProcessOneNodeSync(nodeID)
}

func (c Crawler) MainLoop() {
	crawlerRound := 1
	for {
		startTime := time.Now()
		processedCnt := 0
		done := make(chan bool)
		go func() {
			lastNodeID := ""
			for {
				// get next chunk of nodes from database
				nodesIDs, err := c.DB.GetNextNodesChan(lastNodeID, c.ChunkSize)
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
					nodesLen := c.MaxCoroutineCnt - 1
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
						go c.ProcessOneNodeAsync(lastNodes[i], &wg)
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

package crawling

import (
	"fmt"
	"time"
)

// Crawler handles crawling through each node in OB
type Crawler struct {
	DB           Datastore
	RPCInterface RPCInterface
	WorkerQueue  chan chan Node
}

// CrawlOnce runs the crawler for one step
func (c Crawler) CrawlOnce() (string, error) {
	nextNode, err := c.DB.GetNextNode()
	if err != nil {
		return "", err
	}
	nextNode.LastCrawled = time.Now()

	connections, err := c.RPCInterface.GetConnections(nextNode.ID)

	if err != nil {
		return "", err
	}

	nodes := []Node{}
	for i := range connections {
		if connections[i] != nextNode.ID && len(connections[i]) > 40 && connections[i][0] != ' ' {
			nodes = append(nodes, Node{ID: connections[i], Connections: []string{}, LastCrawled: time.Date(2017, 12, 13, 0, 0, 0, 0, time.Local)})
		}
	}

	nextNode.Connections = connections

	if err := c.DB.AddUninitializedNodes(nodes); err != nil {
		return "", err
	}

	profile, err := c.RPCInterface.GetProfile(nextNode.ID)
	if err != nil {
		return "", err
	}
	if profile.Stats != nil {
		nextNode.Profile = profile

		fmt.Println("saving...")

		if err := c.DB.SaveNode(*nextNode); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("saving uninitialized...")

		if err := c.DB.SaveNodeUninitialized(*nextNode); err != nil {
			fmt.Println(err)
		}
	}
	return nextNode.ID, nil
}

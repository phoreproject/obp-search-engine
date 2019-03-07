package crawling

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Crawler handles crawling through each node in OB
type Crawler struct {
	DB           Datastore
	RPCInterface RPCInterface
	WorkerQueue  chan chan Node
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
		return errors.New(fmt.Sprintf("Could not access node %s. Ignoring.\n  IPNS returned: %s", nextNode.ID, userAgent))
	}
	nextNode.UserAgent = userAgent

	profile, err := c.RPCInterface.GetProfile(nextNode.ID)
	if profile != nil && profile.Stats != nil {
		nextNode.Profile = profile

		fmt.Printf("Saving node %s...\n", nextNode.ID)
		if err := c.DB.SaveNode(nextNode); err != nil {
			return err
		}
	} else {
		fmt.Printf("Saving empty node %s...\n", nextNode.ID)
		if err := c.DB.SaveNodeUninitialized(nextNode); err != nil {
			return err
		}
	}
	return nil
}

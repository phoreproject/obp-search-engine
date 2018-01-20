package crawling

import (
	"log"
	"time"
)

// Datastore represents a way of storing crawled data.
type Datastore interface {
	GetNextNode() (Node, error)
	SaveNode(Node) error
	AddUninitializedNodes([]Node)
	GetNode(string) (Node, error)
}

// Node is a representation of a single node on the network.
type Node struct {
	ID          string
	Connections []string
	LastCrawled time.Time
}

// RPCInterface is an interface to OB
type RPCInterface interface {
	GetConnections(string) ([]string, error)
}

// Crawler handles crawling through each node in OB
type Crawler struct {
	DB           Datastore
	RPCInterface RPCInterface
	WorkerQueue  chan chan Node
}

// CrawlOnce runs the crawler for one step
func (c Crawler) CrawlOnce() error {
	nextNode, err := c.DB.GetNextNode()
	nextNode.LastCrawled = time.Now()

	connections, err := c.RPCInterface.GetConnections(nextNode.ID)

	if err != nil {
		log.Fatal(err)
	}

	nodes := make([]Node, len(connections))
	for i := range connections {
		nodes[i] = Node{ID: connections[i], Connections: []string{}, LastCrawled: time.Date(2017, 12, 13, 0, 0, 0, 0, time.Local)}
	}

	nextNode.Connections = connections

	c.DB.AddUninitializedNodes(nodes)

	c.DB.SaveNode(nextNode)
	return nil
}

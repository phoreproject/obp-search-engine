package crawling

import "time"

// Datastore represents a way of storing crawled data.
type Datastore interface {
	GetNextNode() (*Node, error)
	SaveNode(Node) error
	AddUninitializedNodes([]Node) error
	GetNode(string) (*Node, error)
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

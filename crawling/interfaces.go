package crawling

import (
	"time"
)

// Datastore represents a way of storing crawled data.
type Datastore interface {
	GetNextNode() (*Node, error)
	SaveNode(Node) error
	AddUninitializedNodes([]Node) error
	GetNode(string) (*Node, error)
	AddItemsForNode(owner string, items []Item) error
}

// Node is a representation of a single node on the network.
type Node struct {
	ID          string
	Connections []string
	LastCrawled time.Time
}

// Price is a price in a specific currency
type Price struct {
	CurrencyCode string `json:"currencyCode"`
	Amount       uint64 `json:"amount"`
}

// Thumbnail represents different the addresses for different sizes of thumbnails
type Thumbnail struct {
	Tiny   string `json:"tiny"`
	Small  string `json:"small"`
	Medium string `json:"medium"`
}

// Item represents a single listing in Phore Marketplace
type Item struct {
	Hash          string    `json:"hash"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Categories    []string  `json:"categories"`
	NSFW          bool      `json:"nsfw"`
	ContractType  string    `json:"contractType"`
	Description   string    `json:"description"`
	Thumbnail     Thumbnail `json:"thumbnail"`
	Price         Price     `json:"price"`
	ShipsTo       []string  `json:"shipsTo"`
	FreeShipping  []string  `json:"freeShipping"`
	Language      string    `json:"language"`
	AverageRating float32   `json:"averageRating"`
	RatingCount   uint32    `json:"ratingCount"`
}

// RPCInterface is an interface to OB
type RPCInterface interface {
	GetConnections(string) ([]string, error)
	GetItems(id string) ([]Item, error)
}

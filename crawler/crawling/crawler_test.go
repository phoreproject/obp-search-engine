package crawling_test

import (
	"errors"
	"fmt"
	"github.com/phoreproject/obp-search-engine/crawler/crawling"
	"github.com/phoreproject/obp-search-engine/crawler/mock"
	"testing"
)

func TestCrawlingDBFail(t *testing.T) {
	c := crawling.Crawler{DB: &mock.MockDB{}, RPCInterface: mock.MockRPC{}}

	err := c.CrawlNode()

	if err == nil {
		t.Error(errors.New("Did not error with empty DB."))
	}
}

func TestCrawling(t *testing.T) {
	c := crawling.Crawler{DB: &mock.MockDB{}, RPCInterface: mock.MockRPC{}}

	t.Log("start")

	c.DB.SaveNode(crawling.Node{ID: "1", Connections: []string{}})

	err := c.CrawlNode()

	if err != nil {
		t.Error(err)
	}

	node, err := c.DB.GetNode("1")
	if err != nil {
		t.Error(err)
	}

	if len(node.Connections) != 2 {
		t.Error(fmt.Errorf("Node 1's connections do not match database."))
	}
}

package mock

import (
	"errors"
	"fmt"

	"github.com/phoreproject/obp-search-engine/crawler/crawling"
)

type MockDB struct {
	db []crawling.Node
}

func (m *MockDB) GetNextNode() (*crawling.Node, error) {
	if len(m.db) == 0 {
		return nil, errors.New("database is empty")
	}

	best := 0

	for i := range m.db {
		if len(m.db[i].Connections) == 0 {
			return &m.db[i], nil
		}
		if m.db[best].LastCrawled.Unix() > m.db[i].LastCrawled.Unix() {
			best = i
		}
	}
	return &m.db[best], nil
}

func (m *MockDB) SaveNode(n crawling.Node) error {
	for i := range m.db {
		if m.db[i].ID == n.ID {
			m.db[i] = n
			return nil
		}
	}
	m.db = append(m.db, n)
	return nil
}

func (m *MockDB) AddUninitializedNodes(nodes []crawling.Node) {
	for i := range nodes {
		found := false
		for q := range m.db {
			if nodes[i].ID == m.db[q].ID {
				found = true
				break
			}
		}
		if !found {
			m.SaveNode(nodes[i])
		}
	}
}

func (m *MockDB) GetNode(ID string) (*crawling.Node, error) {
	for i := range m.db {
		if m.db[i].ID == ID {
			return &m.db[i], nil
		}
	}
	return nil, fmt.Errorf("Could not find node with ID %s", ID)
}

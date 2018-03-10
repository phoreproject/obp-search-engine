package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"

	"github.com/phoreproject/obp-search-engine/crawling"
)

// ConfigResponse is the response from the /ob/config RPC command
type ConfigResponse struct {
	Cryptocurrency string `json:"cryptoCurrency"`
	PeerID         string `json:"peerID"`
	Testnet        bool   `json:"testnet"`
	Tor            bool   `json:"tor"`
}

// ClosestPeersResponse is the response from the /ob/closestpeers RPC command
type ClosestPeersResponse []string

// ErrorResponse from RPC
type ErrorResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

// OpenBazaarRPC is a way of sending RPC commands
type OpenBazaarRPC struct {
	URL    string
	client *http.Client
}

// NewRPC creates a new instance of an RPC interface
func NewRPC(url string) OpenBazaarRPC {
	c := &http.Client{}
	return OpenBazaarRPC{URL: url, client: c}
}

// GetConfig gets useful information about the OB client
func (r OpenBazaarRPC) GetConfig() (*ConfigResponse, error) {
	req, err := http.NewRequest("GET", "http://"+path.Join(r.URL, "ob", "config"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	var response ConfigResponse
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetConnections gets connection information about a specific OB client
func (r OpenBazaarRPC) GetConnections(id string) ([]string, error) {
	req, err := http.NewRequest("GET", "http://"+path.Join(r.URL, "ob", "closestpeers", id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	var response ClosestPeersResponse
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetItems gets items for a specific node
func (r OpenBazaarRPC) GetItems(id string) ([]crawling.Item, error) {
	req, err := http.NewRequest("GET", "http://"+path.Join(r.URL, "ob", "listings", id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	responseRaw := new(bytes.Buffer)
	responseRaw.ReadFrom(resp.Body)
	var response []crawling.Item
	if err := json.Unmarshal(responseRaw.Bytes(), &response); err != nil {
		var possibleError ErrorResponse
		if err := json.Unmarshal(responseRaw.Bytes(), &possibleError); err == nil {
			return make([]crawling.Item, 0), nil
		}
		return nil, err
	}
	return response, nil
}

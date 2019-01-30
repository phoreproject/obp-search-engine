package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	ret := []string{}
	for p := range response {
		if response[p][0] != ' ' {
			ret = append(ret, response[p])
		}
	}

	return ret, nil
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
	byteCount, err := responseRaw.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	if byteCount == 0 {
		return nil, nil // fail silently
	}

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

// GetProfile gets the profile for a specific node
func (r OpenBazaarRPC) GetProfile(id string) (*crawling.ProfileResponse, error) {
	req, err := http.NewRequest("GET", "http://"+path.Join(r.URL, "ob", "profile", id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	responseRaw := new(bytes.Buffer)
	byteCount, err := responseRaw.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	if byteCount == 0 {
		return nil, nil // fail silently
	}

	var response crawling.ProfileResponse
	//err = json.Unmarshal(responseRaw.Bytes(), &response) TODO - why duplicated?
	if err := json.Unmarshal(responseRaw.Bytes(), &response); err != nil {
		fmt.Println(err)
		var possibleError ErrorResponse
		if err := json.Unmarshal(responseRaw.Bytes(), &possibleError); err == nil {
			return nil, nil // fail silently
		}
		return nil, err
	}
	return &response, nil
}

// GetUserAgentFromIPNS gets the user agent of a specific node
func (r OpenBazaarRPC) GetUserAgentFromIPNS(id string) (string, error) {
	req, err := http.NewRequest("GET", "http://"+path.Join(r.URL, "ipns", id, "user_agent"), nil)
	if err != nil {
		return "", err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	responseRaw := new(bytes.Buffer)
	_, err = responseRaw.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	var possibleError ErrorResponse
	if err := json.Unmarshal(responseRaw.Bytes(), &possibleError); err == nil {
		return "", errors.New(possibleError.Reason)
	}

	return responseRaw.String(), nil
}

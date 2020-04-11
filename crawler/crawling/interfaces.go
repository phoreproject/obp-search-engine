package crawling

import (
	"time"
)

// Datastore represents a way of storing crawled data.
type Datastore interface {
	GetNextNode() (*Node, error)
	GetNextNodesChan(from string, maxSize int) (<-chan string, error)
	SaveNode(Node) error
	AddUninitializedNodes([]Node) error
	GetNode(string) (*Node, error)
	AddItemsForNode(owner string, items []Item) error
	SaveNodeUninitialized(Node) error
	UpdateNodeStatus(string, string, bool) error
}

// Node is a representation of a single node on the network.
type Node struct {
	ID          string
	UserAgent   string
	Connections []string
	LastCrawled time.Time
	Profile     *ProfileResponse
}

// Price is a price in a specific currency
type Price struct {
	CurrencyCode string `json:"currencyCode"`
	Amount       uint64 `json:"amount"`
	Modifier     uint64 `json:"modifier"`
}

// Thumbnail represents different the addresses for different sizes of thumbnails
type Thumbnail struct {
	Tiny     string `json:"tiny"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Original string `json:"original"`
	Large    string `json:"large"`
}

// Item represents a single listing in Phore Marketplace
type Item struct {
	Score              uint32    `json:"score"` // it is missing in api
	Hash               string    `json:"hash"`
	Slug               string    `json:"slug"`
	Title              string    `json:"title"`
	Tags               []string  `json:"tags"`
	Categories         []string  `json:"categories"`
	NSFW               bool      `json:"nsfw"`
	ContractType       string    `json:"contractType"`
	Format             string    `json:"format"`
	Description        string    `json:"description"`
	Thumbnail          Thumbnail `json:"thumbnail"`
	Price              Price     `json:"price"`
	ShipsTo            []string  `json:"shipsTo"`
	FreeShipping       []string  `json:"freeShipping"`
	Language           string    `json:"language"`
	AverageRating      float32   `json:"averageRating"`
	RatingCount        uint32    `json:"ratingCount"`
	ModeratorIDs       []string  `json:"moderators"`
	AcceptedCurrencies []string  `json:"acceptedCurrencies"`
	CoinType           string    `json:"coinType"`
	CoinDivisibility   uint32    `json:"coinDivisibility"`
	Testnet            bool      `json:"testnet"`
	NormalizedPrice    float64   `json:"normalizedPrice"` // it is missing in api
	Blocked            bool      `json:"blocked"`         // it is missing in api - use only by search engine
	ClassifiedManually bool      `json:"classifiedManually"` // it is missing in api - use only by search engine
}

type ProfileSocialAccount struct {
	Type     string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Username string `protobuf:"bytes,2,opt,name=username" json:"username,omitempty"`
	Proof    string `protobuf:"bytes,3,opt,name=proof" json:"proof,omitempty"`
}

type ProfileContactInfo struct {
	Website     string                  `protobuf:"bytes,1,opt,name=website" json:"website,omitempty"`
	Email       string                  `protobuf:"bytes,2,opt,name=email" json:"email,omitempty"`
	PhoneNumber string                  `protobuf:"bytes,3,opt,name=phoneNumber" json:"phoneNumber,omitempty"`
	Social      []*ProfileSocialAccount `protobuf:"bytes,4,rep,name=social" json:"social,omitempty"`
}

type ProfileColors struct {
	Primary       string `protobuf:"bytes,1,opt,name=primary" json:"primary,omitempty"`
	Secondary     string `protobuf:"bytes,2,opt,name=secondary" json:"secondary,omitempty"`
	Text          string `protobuf:"bytes,3,opt,name=text" json:"text,omitempty"`
	Highlight     string `protobuf:"bytes,4,opt,name=highlight" json:"highlight,omitempty"`
	HighlightText string `protobuf:"bytes,5,opt,name=highlightText" json:"highlightText,omitempty"`
}

type ProfileImage struct {
	Tiny     string `protobuf:"bytes,1,opt,name=tiny" json:"tiny,omitempty"`
	Small    string `protobuf:"bytes,2,opt,name=small" json:"small,omitempty"`
	Medium   string `protobuf:"bytes,3,opt,name=medium" json:"medium,omitempty"`
	Large    string `protobuf:"bytes,4,opt,name=large" json:"large,omitempty"`
	Original string `protobuf:"bytes,5,opt,name=original" json:"original,omitempty"`
}

type ProfileStats struct {
	FollowerCount  uint32  `protobuf:"varint,1,opt,name=followerCount" json:"followerCount,omitempty"`
	FollowingCount uint32  `protobuf:"varint,2,opt,name=followingCount" json:"followingCount,omitempty"`
	ListingCount   uint32  `protobuf:"varint,3,opt,name=listingCount" json:"listingCount,omitempty"`
	RatingCount    uint32  `protobuf:"varint,4,opt,name=ratingCount" json:"ratingCount,omitempty"`
	PostCount      uint32  `protobuf:"varint,5,opt,name=postCount" json:"postCount,omitempty"`
	AverageRating  float32 `protobuf:"fixed32,6,opt,name=averageRating" json:"averageRating,omitempty"`
}

type ModeratorPrice struct {
	CurrencyCode string `protobuf:"bytes,1,opt,name=currencyCode" json:"currencyCode,omitempty"`
	Amount       uint64 `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
}

type ModeratorFee struct {
	FixedFee   *ModeratorPrice `protobuf:"bytes,1,opt,name=fixedFee" json:"fixedFee,omitempty"`
	Percentage float32         `protobuf:"fixed32,2,opt,name=percentage" json:"percentage,omitempty"`
	FeeType    interface{}     `json:"feeType,omitempty"`
}

type Moderator struct {
	Description        string        `protobuf:"bytes,1,opt,name=description" json:"description,omitempty"`
	TermsAndConditions string        `protobuf:"bytes,2,opt,name=termsAndConditions" json:"termsAndConditions,omitempty"`
	Languages          []string      `protobuf:"bytes,3,rep,name=languages" json:"languages,omitempty"`
	AcceptedCurrencies []string      `protobuf:"bytes,4,rep,name=acceptedCurrencies" json:"acceptedCurrencies,omitempty"`
	Fee                *ModeratorFee `protobuf:"bytes,5,opt,name=fee" json:"fee,omitempty"`
}

// ProfileResponse is the response to requests for a user's profile
type ProfileResponse struct {
	PeerID           string              `protobuf:"bytes,1,opt,name=peerID,proto3" json:"peerID,omitempty"`
	Handle           string              `protobuf:"bytes,2,opt,name=handle,proto3" json:"handle,omitempty"`
	Name             string              `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Location         string              `protobuf:"bytes,4,opt,name=location,proto3" json:"location,omitempty"`
	About            string              `protobuf:"bytes,5,opt,name=about,proto3" json:"about,omitempty"`
	ShortDescription string              `protobuf:"bytes,6,opt,name=shortDescription,proto3" json:"shortDescription,omitempty"`
	Nsfw             bool                `protobuf:"varint,7,opt,name=nsfw,proto3" json:"nsfw,omitempty"`
	Vendor           bool                `protobuf:"varint,8,opt,name=vendor,proto3" json:"vendor,omitempty"`
	Moderator        bool                `protobuf:"varint,9,opt,name=moderator,proto3" json:"moderator,omitempty"`
	ModeratorInfo    *Moderator          `protobuf:"bytes,10,opt,name=moderatorInfo,proto3" json:"moderatorInfo,omitempty"`
	ContactInfo      *ProfileContactInfo `protobuf:"bytes,11,opt,name=contactInfo,proto3" json:"contactInfo,omitempty"`
	Colors           *ProfileColors      `protobuf:"bytes,12,opt,name=colors,proto3" json:"colors,omitempty"`
	AvatarHashes     *ProfileImage       `protobuf:"bytes,13,opt,name=avatarHashes,proto3" json:"avatarHashes,omitempty"`
	HeaderHashes     *ProfileImage       `protobuf:"bytes,14,opt,name=headerHashes,proto3" json:"headerHashes,omitempty"`
	Stats            *ProfileStats       `protobuf:"bytes,15,opt,name=stats,proto3" json:"stats,omitempty"`
	BitcoinPubkey    string              `protobuf:"bytes,16,opt,name=bitcoinPubkey,proto3" json:"bitcoinPubkey,omitempty"`
	LastModified     string              `protobuf:"bytes,17,opt,name=lastModified,proto3" json:"lastModified,omitempty"` //in marketplace it is *timestamp.Timestamp from protocol buffer, but it is problematic in parsing to json.
	Currencies       []string            `protobuf:"bytes,18,rep,name=currencies,proto3" json:"currencies,omitempty"`
}

// RPCInterface is an interface to OB
type RPCInterface interface {
	GetConnections(string) ([]string, error)
	GetItems(id string) ([]Item, error)
	GetProfile(id string) (*ProfileResponse, error)
	GetUserAgentFromIPNS(id string) (string, error)
	GetOneItem(guid string, slug string) (*Item, error)
}

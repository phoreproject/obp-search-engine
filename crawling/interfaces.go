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
	SaveNodeUninitialized(Node) error
}

// Node is a representation of a single node on the network.
type Node struct {
	ID          string
	Connections []string
	LastCrawled time.Time
	Profile     *ProfileResponse
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
	FeeType    int32           `protobuf:"varint,3,opt,name=feeType,enum=Moderator_Fee_FeeType" json:"feeType,omitempty"`
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
	PeerID           string              `protobuf:"bytes,1,opt,name=peerID" json:"peerID,omitempty"`
	Handle           string              `protobuf:"bytes,2,opt,name=handle" json:"handle,omitempty"`
	Name             string              `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	Location         string              `protobuf:"bytes,4,opt,name=location" json:"location,omitempty"`
	About            string              `protobuf:"bytes,5,opt,name=about" json:"about,omitempty"`
	ShortDescription string              `protobuf:"bytes,6,opt,name=shortDescription" json:"shortDescription,omitempty"`
	Nsfw             bool                `protobuf:"varint,7,opt,name=nsfw" json:"nsfw,omitempty"`
	Vendor           bool                `protobuf:"varint,8,opt,name=vendor" json:"vendor,omitempty"`
	Moderator        bool                `protobuf:"varint,9,opt,name=moderator" json:"moderator,omitempty"`
	ModeratorInfo    *Moderator          `protobuf:"bytes,10,opt,name=moderatorInfo" json:"moderatorInfo,omitempty"`
	ContactInfo      *ProfileContactInfo `protobuf:"bytes,11,opt,name=contactInfo" json:"contactInfo,omitempty"`
	Colors           *ProfileColors      `protobuf:"bytes,12,opt,name=colors" json:"colors,omitempty"`
	AvatarHashes     *ProfileImage       `protobuf:"bytes,13,opt,name=avatarHashes" json:"avatarHashes,omitempty"`
	HeaderHashes     *ProfileImage       `protobuf:"bytes,14,opt,name=headerHashes" json:"headerHashes,omitempty"`
	Stats            *ProfileStats       `protobuf:"bytes,15,opt,name=stats" json:"stats,omitempty"`
	BitcoinPubkey    string              `protobuf:"bytes,16,opt,name=bitcoinPubkey" json:"bitcoinPubkey,omitempty"`
}

// RPCInterface is an interface to OB
type RPCInterface interface {
	GetConnections(string) ([]string, error)
	GetItems(id string) ([]Item, error)
	GetProfile(id string) (*ProfileResponse, error)
	GetUserAgent(id string) (string, error)
}
